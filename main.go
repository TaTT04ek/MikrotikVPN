package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	mihomoLog "github.com/metacubex/mihomo/log"
)

//go:embed web/*
var webAssets embed.FS

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mihomoLog.SetLevel(mihomoLog.SILENT)
	password := os.Getenv("MIKROTIK_PASSWORD")
	if len(password) < 12 {
		return errors.New("задайте MIKROTIK_PASSWORD длиной от 12 символов")
	}
	dir, err := filepath.Abs(env("MIKROTIK_DATA", "data"))
	if err != nil {
		return err
	}
	coreDir := filepath.Join(dir, "mihomo")
	if err := os.MkdirAll(coreDir, 0700); err != nil {
		return err
	}
	api := coreClient{base: "http://127.0.0.1:9090", secret: rand.Text(), client: &http.Client{
		Timeout: 12 * time.Second, Transport: &http.Transport{MaxIdleConns: 10, MaxIdleConnsPerHost: 10},
	}}
	s, path, err := loadState(dir)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл настроек: %w", err)
	}
	tun := env("MIKROTIK_TUN", "0") == "1"
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	coreDone, err := startCore(ctx, env("MIKROTIK_CORE", "mihomo"), coreDir, s, tun, api)
	if err != nil {
		return err
	}
	m := &manager{current: s, selected: direct, selectedUDP: direct, path: path, core: api, tun: tun, subClient: subscriptionClient(),
		telegramProxy: telegramProxyClient(), probeNow: make(chan struct{}, 1), refreshNow: make(chan struct{}, 1)}
	go m.probeLoop(ctx)
	go m.refreshLoop(ctx)
	go m.telegramLoop(ctx)
	server := &http.Server{Addr: env("MIKROTIK_LISTEN", "127.0.0.1:8080"), Handler: webHandler(m, password),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 65 * time.Second, IdleTimeout: 60 * time.Second}
	webDone := make(chan error, 1)
	go func() { webDone <- server.ListenAndServe() }()
	log.Printf("Веб-панель слушает %s", server.Addr)
	select {
	case err := <-webDone:
		return err
	case err := <-coreDone:
		return fmt.Errorf("Mihomo завершился: %v", err)
	case <-ctx.Done():
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer closeCancel()
		return server.Shutdown(closeCtx)
	}
}

func webHandler(m *manager, password string) http.Handler {
	assets, _ := fs.Sub(webAssets, "web")
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(assets)))
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		status := m.snapshot()
		status.Traffic = readTraffic()
		writeJSON(w, 200, status)
	})
	mux.HandleFunc("GET /api/services", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, m.services()) })
	mux.HandleFunc("POST /api/services/import", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Services []serviceDefinition `json:"services"`
		}
		if readJSON(w, r, &body) {
			result(w, m.importServices(r.Context(), body.Services))
		}
	})
	mux.HandleFunc("POST /api/services/telegram/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !readJSON(w, r, &struct{}{}) {
			return
		}
		result(w, m.refreshTelegram(r.Context()))
	})
	mux.HandleFunc("PUT /api/subscription", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ URL *string }
		if !readJSON(w, r, &body) {
			return
		}
		if body.URL == nil {
			writeJSON(w, 400, map[string]string{"error": "отсутствует поле url"})
			return
		}
		raw := strings.TrimSpace(*body.URL)
		if raw == "" {
			result(w, m.deleteSource(r.Context(), legacySourceID))
			return
		}
		result(w, m.addSource(r.Context(), legacySourceID, raw))
	})
	mux.HandleFunc("POST /api/subscription/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !readJSON(w, r, &struct{}{}) {
			return
		}
		result(w, m.refreshSource(r.Context(), legacySourceID))
	})
	mux.HandleFunc("POST /api/sources", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Input string }
		if !readJSON(w, r, &body) {
			return
		}
		result(w, m.addSource(r.Context(), "", body.Input))
	})
	mux.HandleFunc("POST /api/sources/{id}/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !readJSON(w, r, &struct{}{}) {
			return
		}
		result(w, m.refreshSource(r.Context(), r.PathValue("id")))
	})
	mux.HandleFunc("DELETE /api/sources/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !readJSON(w, r, &struct{}{}) {
			return
		}
		result(w, m.deleteSource(r.Context(), r.PathValue("id")))
	})
	mux.HandleFunc("PUT /api/settings", func(w http.ResponseWriter, r *http.Request) {
		var body settings
		if !readJSON(w, r, &body) {
			return
		}
		result(w, m.saveSettings(r.Context(), body))
	})
	mux.HandleFunc("POST /api/probe", func(w http.ResponseWriter, r *http.Request) {
		if !readJSON(w, r, &struct{}{}) {
			return
		}
		notify(m.probeNow)
		w.WriteHeader(http.StatusAccepted)
	})
	passwordHash := sha256.Sum256([]byte(password))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		user, pass, ok := r.BasicAuth()
		hash := sha256.Sum256([]byte(pass))
		if !ok || user != "admin" || subtle.ConstantTimeCompare(hash[:], passwordHash[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Mikrot VPN", charset="UTF-8"`)
			http.Error(w, "Требуется вход", http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	kind, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if kind != "application/json" {
		writeJSON(w, 415, map[string]string{"error": "требуется application/json"})
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if decoder.Decode(value) != nil || decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, 400, map[string]string{"error": "некорректное тело запроса"})
		return false
	}
	return true
}

func result(w http.ResponseWriter, err error) {
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}
