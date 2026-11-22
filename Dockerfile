# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.27.4-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
WORKDIR /src
RUN apk add --no-cache ca-certificates curl
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY web ./web
COPY services ./services
RUN set -eu; \
    case "$TARGETARCH/$TARGETVARIANT" in \
      arm/v5|arm/v6|arm/v7) export GOARM="${TARGETVARIANT#v},softfloat"; asset="arm$TARGETVARIANT" ;; \
      arm64/*) asset=arm64 ;; \
      amd64/*) asset=amd64-v1 ;; \
      *) exit 1 ;; \
    esac; \
    case "$asset" in \
      armv5) sha=f0e90b3cdd957c8e8bd6e5ff515d7bf310452734af15c90fdda4f6fc212689a0 ;; \
      armv6) sha=efa575b877b1c177e261492f994e4ec0d9bd6b92deed2015e25612396a9ce487 ;; \
      armv7) sha=f622cbe903ba1a11d30e9071d1c5373929c675b854856ff3e7f977255a25bb66 ;; \
      arm64) sha=9dd862e28b46ff7d775f169cceebc28deccaa0a9e804237d421cd2571e0caba0 ;; \
      amd64-v1) sha=306f81e723e60ce6b828899a6fe83e1d00e9ecefb2dc8d4d849312a5bc00efdc ;; \
    esac; \
    curl -fL --retry 3 "https://github.com/MetaCubeX/mihomo/releases/download/v1.19.32/mihomo-linux-$asset-v1.19.32.gz" -o /tmp/mihomo.gz; \
    echo "$sha  /tmp/mihomo.gz" | sha256sum -c -; \
    mkdir -p /out; \
    gunzip -c /tmp/mihomo.gz > /out/mihomo; \
    chmod 755 /out/mihomo; \
    curl -fL --retry 3 "https://raw.githubusercontent.com/MetaCubeX/mihomo/v1.19.32/LICENSE" -o /out/Mihomo-LICENSE; \
    CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" go build -trimpath -ldflags="-s -w" -o /out/mikrot .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/mikrot /out/mihomo /bin/
COPY --from=build /out/Mihomo-LICENSE /usr/share/licenses/mihomo/LICENSE
COPY --from=build /src/services/LICENSE /usr/share/licenses/domain-list-community/LICENSE
ENV MIKROTIK_CORE=/bin/mihomo MIKROTIK_DATA=/data MIKROTIK_LISTEN=0.0.0.0:8080 MIKROTIK_TUN=1 GOMEMLIMIT=48MiB
EXPOSE 8080
ENTRYPOINT ["/bin/mikrot"]
