// request_mapping_test.go verifies that CLI flags are translated into the
// exact JSON bodies the backend expects. This is where the impedance mismatch
// between short CLI flag names and the backend's JSON field names is caught.

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// readBody parses the request body into a generic map for field inspection.
func readBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	return m
}

func checkKey(t *testing.T, body map[string]any, key string, want any) {
	t.Helper()
	got, ok := body[key]
	if !ok {
		t.Errorf("JSON body missing key %q", key)
		return
	}
	if got != want {
		t.Errorf("body[%q] = %v, want %v", key, got, want)
	}
}

func checkAbsent(t *testing.T, body map[string]any, key string) {
	t.Helper()
	if _, ok := body[key]; ok {
		t.Errorf("JSON body must not contain key %q (omitempty)", key)
	}
}

// ── gateways create ───────────────────────────────────────────────────────────

func TestGatewaysCreate_FlagToJSONMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sim/gateways" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body := readBody(t, r)
		checkKey(t, body, "factoryId", "fac-42")           // --factory-id
		checkKey(t, body, "factoryKey", "secret-key")      // --factory-key
		checkKey(t, body, "serialNumber", "SN-XYZ")        // --serial
		checkKey(t, body, "model", "GW-PRO")               // --model
		checkKey(t, body, "firmwareVersion", "3.0.1")      // --firmware
		checkKey(t, body, "sendFrequencyMs", float64(250)) // --freq

		writeJSON(w, http.StatusCreated, map[string]any{"id": 1})
	})

	err := runCmd("gateways", "create",
		"--factory-id", "fac-42",
		"--factory-key", "secret-key",
		"--serial", "SN-XYZ",
		"--model", "GW-PRO",
		"--firmware", "3.0.1",
		"--freq", "250",
	)
	if err != nil {
		t.Fatalf("gateways create failed: %v", err)
	}
}

func TestGatewaysCreate_DefaultFreq_Is1000(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		// --freq defaults to 1000 ms when not provided
		checkKey(t, body, "sendFrequencyMs", float64(1000))
		writeJSON(w, http.StatusCreated, map[string]any{"id": 1})
	})

	err := runCmd("gateways", "create",
		"--factory-id", "f",
		"--factory-key", "k",
		"--serial", "SN",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGatewaysCreate_OptionalFields_OmittedWhenNotProvided(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		// model and firmware are optional; when not passed they must be absent
		// (omitempty), not sent as empty strings
		checkAbsent(t, body, "model")
		checkAbsent(t, body, "firmwareVersion")
		writeJSON(w, http.StatusCreated, map[string]any{"id": 1})
	})

	err := runCmd("gateways", "create",
		"--factory-id", "f",
		"--factory-key", "k",
		"--serial", "SN",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── gateways bulk ─────────────────────────────────────────────────────────────

func TestGatewaysBulk_FlagToJSONMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sim/gateways/bulk" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body := readBody(t, r)
		checkKey(t, body, "count", float64(5))
		checkKey(t, body, "factoryId", "fac-bulk")
		checkKey(t, body, "factoryKey", "key-bulk")
		checkKey(t, body, "model", "GW-MINI")
		checkKey(t, body, "firmwareVersion", "1.2.3")
		checkKey(t, body, "sendFrequencyMs", float64(500))

		writeJSON(w, http.StatusCreated, map[string]any{
			"gateways": []any{},
			"errors":   []any{},
		})
	})

	err := runCmd("gateways", "bulk",
		"--count", "5",
		"--factory-id", "fac-bulk",
		"--factory-key", "key-bulk",
		"--model", "GW-MINI",
		"--firmware", "1.2.3",
		"--freq", "500",
	)
	if err != nil {
		t.Fatalf("gateways bulk failed: %v", err)
	}
}

// ── gateways get ─────────────────────────────────────────────────────────────

func TestGatewaysGet_Integration(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/sim/gateways/uuid-get-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                  42,
			"managementGatewayId": "uuid-get-1",
			"status":              "online",
			"model":               "GW-X",
			"serialNumber":        "SN001",
			"firmwareVersion":     "1.0",
			"sendFrequencyMs":     1000,
			"tenantId":            "t-1",
			"provisioned":         true,
			"createdAt":           "2024-01-01T00:00:00Z",
		})
	})

	if err := runCmd("gateways", "get", "uuid-get-1"); err != nil {
		t.Fatalf("gateways get failed: %v", err)
	}
}

func TestGatewaysGet_NotFound(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	if err := runCmd("gateways", "get", "ghost-uuid"); err == nil {
		t.Error("expected error on 404")
	}
}

// ── sensors add ───────────────────────────────────────────────────────────────
//
// CRITICAL: --min/--max CLI flags must map to "minRange"/"maxRange" in JSON.
// If this mapping breaks the backend will reject the request or use wrong defaults.

func TestSensorsAdd_FlagToJSONMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sim/gateways/5/sensors" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body := readBody(t, r)

		// --type → "type"
		checkKey(t, body, "type", "humidity")
		// --min → "minRange" (NOT "min")
		checkKey(t, body, "minRange", float64(10))
		checkAbsent(t, body, "min")
		// --max → "maxRange" (NOT "max")
		checkKey(t, body, "maxRange", float64(90))
		checkAbsent(t, body, "max")
		// --algorithm → "algorithm"
		checkKey(t, body, "algorithm", "uniform_random")

		writeJSON(w, http.StatusCreated, map[string]any{
			"id": 1, "gatewayId": 5, "sensorId": "s-uuid",
			"type": "humidity", "minRange": 10, "maxRange": 90, "algorithm": "uniform_random",
		})
	})

	err := runCmd("sensors", "add", "5",
		"--type", "humidity",
		"--min", "10",
		"--max", "90",
		"--algorithm", "uniform_random",
	)
	if err != nil {
		t.Fatalf("sensors add failed: %v", err)
	}
}

func TestSensorsAdd_NegativeMinRange(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		checkKey(t, body, "minRange", float64(-40))
		checkKey(t, body, "maxRange", float64(85))
		writeJSON(w, http.StatusCreated, map[string]any{"id": 2, "gatewayId": 3, "sensorId": "s-2"})
	})

	err := runCmd("sensors", "add", "3",
		"--type", "temperature",
		"--min", "-40",
		"--max", "85",
		"--algorithm", "sine_wave",
	)
	if err != nil {
		t.Fatalf("sensors add with negative min failed: %v", err)
	}
}

// ── anomalies disconnect ──────────────────────────────────────────────────────

func TestAnomaliesDisconnect_DurationFieldName(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/gateways/uuid-dc/anomaly/disconnect" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		// Backend expects "duration_seconds", not "duration"
		checkKey(t, body, "duration_seconds", float64(10))
		checkAbsent(t, body, "duration")
		checkAbsent(t, body, "durationSeconds")
		w.WriteHeader(http.StatusNoContent)
	})

	if err := runCmd("anomalies", "disconnect", "uuid-dc", "--duration", "10"); err != nil {
		t.Fatalf("anomalies disconnect failed: %v", err)
	}
}

// ── anomalies network-degradation ─────────────────────────────────────────────

func TestAnomaliesNetworkDegradation_FieldNames(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		checkKey(t, body, "duration_seconds", float64(20))
		checkKey(t, body, "packet_loss_pct", 0.5)
		checkAbsent(t, body, "packetLoss")
		w.WriteHeader(http.StatusNoContent)
	})

	err := runCmd("anomalies", "network-degradation", "uuid-nd",
		"--duration", "20",
		"--packet-loss", "0.5",
	)
	if err != nil {
		t.Fatalf("anomalies network-degradation failed: %v", err)
	}
}

func TestAnomaliesNetworkDegradation_PacketLossOmitted_WhenNotProvided(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		// --packet-loss not passed → field absent → backend uses its default (0.3)
		checkAbsent(t, body, "packet_loss_pct")
		w.WriteHeader(http.StatusNoContent)
	})

	err := runCmd("anomalies", "network-degradation", "uuid-nd", "--duration", "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── anomalies outlier ─────────────────────────────────────────────────────────

func TestAnomaliesOutlier_ValueOmitted_WhenNotProvided(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		// --value not passed → field must be absent (nil pointer + omitempty)
		checkAbsent(t, body, "value")
		w.WriteHeader(http.StatusNoContent)
	})

	if err := runCmd("anomalies", "outlier", "10"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
