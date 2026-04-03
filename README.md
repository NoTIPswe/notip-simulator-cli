# sim-cli — NoTIP Simulator CLI

A Go-based CLI for managing the NoTIP Simulator: gateways, sensors, and anomaly triggers.

## Requirements

- Docker & Docker Compose (no host-machine Go installation needed)

## Usage

The CLI is packaged as an ephemeral Docker container. It is spun up on demand, runs one command, and is immediately destroyed.

### Help

```bash
docker compose run --rm sim-cli --help
```

### Gateways

```bash
# List all gateways (requires -it for styled output)
docker compose run --rm -it sim-cli gateways list

# Create 5 gateways for a tenant
docker compose run --rm sim-cli gateways create --count 5 --tenant <uuid>

# Delete a gateway by ID
docker compose run --rm sim-cli gateways delete <id>
```

### Sensors

```bash
# Add a temperature sensor to a gateway
docker compose run --rm sim-cli sensors add <gateway-id> --type temperature --min 20.0 --max 80.0
```

### Anomalies

```bash
# Trigger a disconnect anomaly on a gateway
docker compose run --rm sim-cli anomalies disconnect <gateway-id>
```

## Docker Compose configuration

Add the following service to your `docker-compose.yml`:

```yaml
  sim-cli:
    image: ghcr.io/notipswe/notip-sim-cli:latest
    profiles:
      - cli
    environment:
      SIMULATOR_URL: http://simulator:8090
    networks:
      - internal
```

The `cli` profile ensures the container never starts automatically. The `SIMULATOR_URL` env var can be overridden to target a different backend.

## TTY awareness

When run without `-it` (e.g. in scripts or CI), PTerm styling and ANSI colours are automatically disabled and output falls back to plain text.
