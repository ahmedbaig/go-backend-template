# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.27
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates tzdata
COPY . .
ARG CMD=./cmd/server
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ${CMD}

FROM alpine:3.22
RUN apk add --no-cache tini ca-certificates tzdata wget \
 && addgroup -S -g 10001 app && adduser -S -u 10001 -G app -h /app app \
 && mkdir -p /app/var/secrets /app/var/logs && chown -R 10001:10001 /app/var
ARG TZ=Australia/Sydney
ENV TZ=${TZ} CONFIG_DIR=/app/var/secrets PORT=8080
WORKDIR /app
COPY --from=build /out/app /app/app
USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/bin/sh", "-c", "wget -qO- http://127.0.0.1:${PORT}/healthz >/dev/null || exit 1"]
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/app"]
