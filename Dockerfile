# ─── Builder (CI/CD) ────────────────────────────────────────────────────────
FROM ghcr.io/notipswe/notip-go-base:v0.0.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o sim-cli .

# ─── Production ─────────────────────────────────────────────────────────────
FROM ghcr.io/notipswe/notip-go-base:v0.0.1 AS prod

LABEL org.opencontainers.image.source="https://github.com/NoTIPswe/notip-simulator-cli" \
      org.opencontainers.image.description="NoTIP Simulator CLI" \
      org.opencontainers.image.licenses="MIT"

RUN groupadd -r appuser && useradd -r -g appuser appuser \
    && apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --chown=appuser:appuser --from=builder /app/sim-cli .

USER appuser

ENTRYPOINT ["./sim-cli"]
CMD ["--help"]
