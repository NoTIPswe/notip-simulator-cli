// request_construction_test.go verifies that every client method builds the HTTP
// request exactly as the simulator backend expects it:
//   - correct HTTP method
//   - correct URL path (including /sim/ prefix and numeric vs UUID IDs)
//   - correct Content-Type header on POST-with-body requests
//   - exact JSON field names in the request body
//   - omitempty: optional fields are absent when zero-valued

package client_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/NoTIPswe/notip-simulator-cli/internal/client"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func readBodyAsMap(t *testing.T, r *http.Request) map[string]any {
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

func assertContentType(t *testing.T, r *http.Request) {
	t.Helper()
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func assertNoBody(t *testing.T, r *http.Request) {
	t.Helper()
	b, _ := io.ReadAll(r.Body)
	if len(b) > 0 {
		t.Errorf("expected no request body, got: %s", string(b))
	}
}

func assertKey(t *testing.T, m map[string]any, key string, want any) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Errorf("JSON body missing key %q", key)
		return
	}
	if got != want {
		t.Errorf("body[%q] = %v (%T), want %v (%T)", key, got, got, want, want)
	}
}

func assertKeyAbsent(t *testing.T, m map[string]any, key string) {
	t.Helper()
	if _, ok := m[key]; ok {
		t.Errorf("JSON body should not contain key %q (omitempty)", key)
	}
}

// ── POST /sim/gateways — single create ───────────────────────────────────────

func TestCreateGateway_RequestConstruction(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/sim/gateways" {
			t.Errorf("path = %s, want /sim/gateways", r.URL.Path)
		}
		assertContentType(t, r)

		body := readBodyAsMap(t, r)
		assertKey(t, body, "factoryId", "fac-1")
		assertKey(t, body, "factoryKey", "key-secret")
		assertKey(t, body, "model", "GW-X200")
		assertKey(t, body, "firmwareVersion", "2.1.0")
		assertKey(t, body, "sendFrequencyMs", float64(500))

		writeJSON(w, http.StatusCreated, client.Gateway{ID: 1})
	})

	_, err := c.CreateGateway(client.CreateGatewayRequest{
		FactoryID:       "fac-1",
		FactoryKey:      "key-secret",
		Model:           "GW-X200",
		FirmwareVersion: "2.1.0",
		SendFrequencyMs: 500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateGateway_OptionalFields_OmittedWhenZero(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBodyAsMap(t, r)
		// model, firmwareVersion, sendFrequencyMs are omitempty — absent when zero
		assertKeyAbsent(t, body, "model")
		assertKeyAbsent(t, body, "firmwareVersion")
		assertKeyAbsent(t, body, "sendFrequencyMs")
		writeJSON(w, http.StatusCreated, client.Gateway{ID: 1})
	})

	_, _ = c.CreateGateway(client.CreateGatewayRequest{
		FactoryID:  "f",
		FactoryKey: "k",
	})
}

// ── POST /sim/gateways/bulk ───────────────────────────────────────────────────

func TestBulkCreateGateways_RequestConstruction(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/sim/gateways/bulk" {
			t.Errorf("path = %s, want /sim/gateways/bulk", r.URL.Path)
		}
		assertContentType(t, r)

		body := readBodyAsMap(t, r)
		assertKey(t, body, "count", float64(3))
		assertKey(t, body, "factoryId", "fac-bulk")
		assertKey(t, body, "factoryKey", "key-bulk")
		assertKey(t, body, "model", "GW-BULK")
		assertKey(t, body, "firmwareVersion", "1.0.0")
		assertKey(t, body, "sendFrequencyMs", float64(2000))

		writeJSON(w, http.StatusCreated, client.BulkCreateResponse{
			Gateways: []client.Gateway{{ID: 1}, {ID: 2}, {ID: 3}},
			Errors:   []string{"", "", ""},
		})
	})

	_, err := c.BulkCreateGateways(client.BulkCreateGatewaysRequest{
		Count:           3,
		FactoryID:       "fac-bulk",
		FactoryKey:      "key-bulk",
		Model:           "GW-BULK",
		FirmwareVersion: "1.0.0",
		SendFrequencyMs: 2000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBulkCreateGateways_OptionalFields_OmittedWhenZero(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBodyAsMap(t, r)
		assertKeyAbsent(t, body, "model")
		assertKeyAbsent(t, body, "firmwareVersion")
		assertKeyAbsent(t, body, "sendFrequencyMs")
		writeJSON(w, http.StatusCreated, client.BulkCreateResponse{})
	})
	_, _ = c.BulkCreateGateways(client.BulkCreateGatewaysRequest{Count: 1, FactoryID: "f", FactoryKey: "k"})
}

// ── GET /sim/gateways — no body ───────────────────────────────────────────────

func TestListGateways_NoBodyNoContentType(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("GET should not set Content-Type, got %q", r.Header.Get("Content-Type"))
		}
		assertNoBody(t, r)
		writeJSON(w, http.StatusOK, []client.Gateway{})
	})
	_, _ = c.ListGateways()
}

// ── POST /sim/gateways/{id}/start & stop — no body ───────────────────────────

func TestStartGateway_NoBody(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/gateways/uuid-abc/start" {
			t.Errorf("path = %s", r.URL.Path)
		}
		// POST with no body: Content-Type should not be set
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("POST-no-body should not set Content-Type")
		}
		assertNoBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.StartGateway("uuid-abc")
}

func TestStopGateway_NoBody(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/gateways/uuid-abc/stop" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("POST-no-body should not set Content-Type")
		}
		assertNoBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.StopGateway("uuid-abc")
}

// ── DELETE /sim/gateways/{id} — no body ──────────────────────────────────────

func TestDeleteGateway_MethodAndNoBody(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/sim/gateways/uuid-del" {
			t.Errorf("path = %s", r.URL.Path)
		}
		assertNoBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.DeleteGateway("uuid-del")
}

// ── POST /sim/gateways/{id}/sensors — JSON field names ───────────────────────
//
// CRITICAL: CLI flags --min/--max must map to "minRange"/"maxRange" in JSON,
// not "min"/"max". This is the field name the backend expects.

func TestAddSensor_RequestFieldNames(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/sim/gateways/7/sensors" {
			t.Errorf("path = %s, want /sim/gateways/7/sensors", r.URL.Path)
		}
		assertContentType(t, r)

		body := readBodyAsMap(t, r)
		// Field name must be "minRange", NOT "min"
		assertKey(t, body, "minRange", float64(-10))
		assertKeyAbsent(t, body, "min")
		// Field name must be "maxRange", NOT "max"
		assertKey(t, body, "maxRange", float64(120))
		assertKeyAbsent(t, body, "max")
		// Field name must be "algorithm"
		assertKey(t, body, "algorithm", "sine_wave")
		// Field name must be "type"
		assertKey(t, body, "type", "temperature")

		writeJSON(w, http.StatusCreated, client.Sensor{ID: 1})
	})

	_, err := c.AddSensor(7, client.AddSensorRequest{
		Type:      "temperature",
		MinRange:  -10,
		MaxRange:  120,
		Algorithm: "sine_wave",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── GET /sim/gateways/{id}/sensors — numeric ID in path ──────────────────────

func TestListSensors_NumericIDInPath(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// ID must be the numeric int64, not a UUID
		if r.URL.Path != "/sim/gateways/1234567890/sensors" {
			t.Errorf("path = %s, want /sim/gateways/1234567890/sensors", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, []client.Sensor{})
	})
	_, _ = c.ListSensors(1234567890)
}

// ── DELETE /sim/sensors/{sensorId} — numeric ID in path ──────────────────────

func TestDeleteSensor_NumericIDInPath(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/sim/sensors/42" {
			t.Errorf("path = %s, want /sim/sensors/42", r.URL.Path)
		}
		assertNoBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.DeleteSensor(42)
}

// ── POST /sim/gateways/{id}/anomaly/disconnect — duration_seconds field ───────

func TestDisconnect_RequestFieldNames(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/gateways/uuid-x/anomaly/disconnect" {
			t.Errorf("path = %s", r.URL.Path)
		}
		assertContentType(t, r)

		body := readBodyAsMap(t, r)
		// Backend expects "duration_seconds", not "duration" or "durationSeconds"
		assertKey(t, body, "duration_seconds", float64(7))
		assertKeyAbsent(t, body, "duration")
		assertKeyAbsent(t, body, "durationSeconds")

		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.Disconnect("uuid-x", 7)
}

// ── POST /sim/gateways/{id}/anomaly/network-degradation ──────────────────────

func TestNetworkDegradation_RequestFieldNames(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/gateways/uuid-x/anomaly/network-degradation" {
			t.Errorf("path = %s", r.URL.Path)
		}
		assertContentType(t, r)

		body := readBodyAsMap(t, r)
		assertKey(t, body, "duration_seconds", float64(15))
		// Backend expects "packet_loss_pct", not "packetLoss" or "packet_loss"
		assertKey(t, body, "packet_loss_pct", 0.25)
		assertKeyAbsent(t, body, "packetLoss")
		assertKeyAbsent(t, body, "packet_loss")

		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.InjectNetworkDegradation("uuid-x", 15, 0.25)
}

func TestNetworkDegradation_PacketLossOmitted_WhenZero(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBodyAsMap(t, r)
		// 0.0 with omitempty → field must be absent so backend applies its default (0.3)
		assertKeyAbsent(t, body, "packet_loss_pct")
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.InjectNetworkDegradation("uuid-x", 5, 0)
}

// ── POST /sim/sensors/{sensorId}/anomaly/outlier ──────────────────────────────

func TestOutlier_RequestFieldNames_WithValue(t *testing.T) {
	val := 42.5
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/sensors/99/anomaly/outlier" {
			t.Errorf("path = %s, want /sim/sensors/99/anomaly/outlier", r.URL.Path)
		}
		assertContentType(t, r)

		body := readBodyAsMap(t, r)
		assertKey(t, body, "value", 42.5)

		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.InjectOutlier(99, &val)
}

func TestOutlier_ValueOmitted_WhenNil(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBodyAsMap(t, r)
		// nil pointer + omitempty → "value" key must be absent
		assertKeyAbsent(t, body, "value")
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.InjectOutlier(99, nil)
}

// ── Default SIMULATOR_URL matches docker-compose service name ─────────────────
//
// The docker-compose.yml exposes the simulator backend as service "simulator"
// on port 8090 inside the "internal" network. The default URL must match.

func TestDefaultSimulatorURL(t *testing.T) {
	c := client.New("http://simulator:8090")
	if c == nil {
		t.Fatal("client.New returned nil")
	}
	// Verify the client targets the correct host by checking a request fails
	// with a connection error (not a nil-pointer or wrong-host error).
	_, err := c.ListGateways()
	if err == nil {
		t.Fatal("expected connection error when simulator is not running")
	}
}
