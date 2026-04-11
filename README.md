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

---

### Gateways

```bash
# List all gateways (requires -it for styled output)
docker compose run --rm -it sim-cli gateways list

# Show details for a single gateway
docker compose run --rm -it sim-cli gateways get <gateway-uuid>

# Create a single gateway
docker compose run --rm sim-cli gateways create \
  --factory-id FAC-001 \
  --factory-key KEY-001 \
  --model GW-X \
  --firmware 1.0.0 \
  --freq 1000

# Bulk create gateways with one shared config and multiple factory IDs
docker compose run --rm sim-cli gateways bulk \
  --factory-id FAC-001 \
  --factory-id FAC-002 \
  --factory-id FAC-003 \
  --factory-key KEY-001 \
  --model GW-X \
  --firmware 1.0.0 \
  --freq 1000

# Start telemetry emission for a gateway
docker compose run --rm sim-cli gateways start <gateway-uuid>

# Stop telemetry emission for a gateway
docker compose run --rm sim-cli gateways stop <gateway-uuid>

# Delete a gateway by UUID
docker compose run --rm sim-cli gateways delete <gateway-uuid>
```

---

### Sensors

```bash
# List all sensors for a gateway
docker compose run --rm -it sim-cli sensors list <gateway-uuid>

# Add a sensor to a gateway
docker compose run --rm sim-cli sensors add <gateway-uuid> \
  --type temperature \
  --min 20.0 \
  --max 80.0 \
  --algorithm uniform_random

# Delete a sensor by UUID
docker compose run --rm sim-cli sensors delete <sensor-uuid>
```

**Sensor types:** `temperature` | `humidity` | `pressure` | `movement` | `biometric`

**Generation algorithms:** `uniform_random` | `sine_wave` | `spike` | `constant`

---

### Anomalies

```bash
# Simulate a gateway disconnect for a given duration (seconds)
docker compose run --rm sim-cli anomalies disconnect <gateway-uuid> --duration 10

# Simulate network degradation (packet loss) on a gateway
docker compose run --rm sim-cli anomalies network-degradation <gateway-uuid> \
  --duration 30 \
  --packet-loss 0.3   # fraction 0–1; omit to use backend default (0.3)

# Inject an outlier reading into a sensor
docker compose run --rm sim-cli anomalies outlier <sensor-uuid>
docker compose run --rm sim-cli anomalies outlier <sensor-uuid> --value 999.9
```

---

### Interactive shell

Start a persistent REPL session to run multiple commands without restarting the container:

```bash
docker compose run --rm -it sim-cli shell
```

Inside the shell, type commands directly (without the `sim-cli` prefix):

```
sim-cli> gateways list
sim-cli> sensors list <gateway-uuid>
sim-cli> anomalies disconnect <uuid> --duration 5
sim-cli> exit
```

Type `exit` or press `Ctrl+D` to quit.

---

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
