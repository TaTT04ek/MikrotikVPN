package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type traffic struct {
	Interface string `json:"interface"`
	Received  uint64 `json:"received"`
	Sent      uint64 `json:"sent"`
	SampledAt int64  `json:"sampledAt"`
}

func readTraffic() *traffic {
	routes, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return nil
	}
	devices, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil
	}
	return interfaceTraffic(string(routes), string(devices), time.Now().UnixMilli())
}

func interfaceTraffic(routes, devices string, now int64) *traffic {
	// Count the uplink only. Including the tunnel would count some packets twice.
	var uplink string
	var bestMetric uint64
	for _, line := range strings.Split(routes, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 8 || fields[1] != "00000000" || fields[7] != "00000000" {
			continue
		}
		flags, flagErr := strconv.ParseUint(fields[3], 16, 32)
		metric, metricErr := strconv.ParseUint(fields[6], 10, 64)
		if flagErr == nil && metricErr == nil && flags&1 != 0 && (uplink == "" || metric < bestMetric) {
			uplink, bestMetric = fields[0], metric
		}
	}
	if uplink == "" {
		return nil
	}
	for _, line := range strings.Split(devices, "\n") {
		name, values, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(name) != uplink {
			continue
		}
		fields := strings.Fields(values)
		if len(fields) < 16 {
			return nil
		}
		received, rxErr := strconv.ParseUint(fields[0], 10, 64)
		sent, txErr := strconv.ParseUint(fields[8], 10, 64)
		if rxErr == nil && txErr == nil {
			return &traffic{Interface: uplink, Received: received, Sent: sent, SampledAt: now}
		}
	}
	return nil
}
