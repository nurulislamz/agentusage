# syntax=docker/dockerfile:1
#
# Dockerfile — multi-stage image for the unified agentusage binary.
#
# Scope: bundles the unified `agentusage` CLI binary. Can be run in headless
# hub mode (`agentusage hub --headless`), web dashboard mode (`agentusage serve`),
# telemetry daemon mode (`agentusage telemetry daemon run`), or export modes in a container.
#
# CGO is enabled for mattn/go-sqlite3 (used by telemetry storage & Cursor provider).

# ── builder ──────────────────────────────────────────────────────────────────
FROM golang:1.25-alpine3.21 AS builder

# CGO is required for mattn/go-sqlite3.
RUN apk add --no-cache gcc musl-dev git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-s -w \
      -X 'github.com/nurulislamz/agentusage/internal/version.Version=${VERSION}' \
      -X 'github.com/nurulislamz/agentusage/internal/version.CommitHash=${COMMIT_HASH}' \
      -X 'github.com/nurulislamz/agentusage/internal/version.BuildDate=${BUILD_DATE}'" \
    -o /agentusage ./cmd/agentusage

# ── runtime ───────────────────────────────────────────────────────────────────
FROM alpine:3.21

# ca-certificates: HTTPS calls to provider APIs.
# tzdata: timezone handling in logs and timestamps.
# wget, curl: minimal HTTP clients for HEALTHCHECK / scripts.
RUN apk add --no-cache ca-certificates tzdata wget curl

COPY --from=builder /agentusage /usr/local/bin/agentusage

# OCI image labels. Values are filled by the release pipeline build args.
ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_DATE=unknown
LABEL org.opencontainers.image.title="agentusage" \
      org.opencontainers.image.description="agentUsage: terminal-first local quota and usage tracking for AI coding agents, IDEs, and LLM APIs." \
      org.opencontainers.image.source="https://github.com/nurulislamz/agentusage" \
      org.opencontainers.image.url="https://github.com/nurulislamz/agentusage" \
      org.opencontainers.image.documentation="https://github.com/nurulislamz/agentusage" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT_HASH}" \
      org.opencontainers.image.created="${BUILD_DATE}"

# Create unprivileged agentusage user (UID 1000) and directories for config and state/socket persistence.
RUN addgroup -g 1000 agentusage && \
    adduser -u 1000 -G agentusage -h /home/agentusage -s /bin/sh -D agentusage && \
    mkdir -p /home/agentusage/.config/agentusage \
             /home/agentusage/.local/state/agentusage \
             /home/agentusage/.local/share/agentusage && \
    chown -R agentusage:agentusage /home/agentusage

ENV HOME=/home/agentusage
USER 1000:1000
WORKDIR /home/agentusage

# 9190 for hub server, 8080 for web serve dashboard
EXPOSE 9190 8080

# HEALTHCHECK verifies either the hub health endpoint (:9190/healthz) or the web dashboard health endpoint (:8080/healthz)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:9190/healthz || wget --quiet --tries=1 --spider http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["agentusage"]
CMD ["hub", "--headless"]
