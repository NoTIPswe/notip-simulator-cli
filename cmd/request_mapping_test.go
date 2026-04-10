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

const (
	fmtUnexpectedError = "unexpected error: %v"
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

// -- gateways create ----------------------------------------------------------

func TestGatewaysCreateFlagToJSONMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sim/gateways" {
			t.Errorf(fmtUnexpectedRequest, r.Method, r.URL.Path)
		}
		body := readBody(t, r)
		checkKey(t, body, "factoryId", "fac-42")
		checkKey(t, body, "factoryKey", "secret-key")
		checkKey(t, body, "model", "GW-PRO")
		checkKey(t, body, "firmwareVersion", "3.0.1")
		checkKey(t, body, "sendFrequencyMs", float64(250))

		writeJSON(w, http.StatusCreated, map[string]any{"id": "gw-1"})
	})

	err := runCmd("gateways", "create",
		testFlagFactoryID, "fac-42",
		testFlagFactoryKey, "secret-key",
		testFlagModel, "GW-PRO",
		testFlagFirmware, "3.0.1",
		testFlagFreq, "250",
	)
	if err != nil {
		t.Fatalf("gateways create failed: %v", err)
	}
}

func TestGatewaysCreateFreqMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		checkKey(t, body, "sendFrequencyMs", float64(1000))
		writeJSON(w, http.StatusCreated, map[string]any{"id": "gw-1"})
	})

	err := runCmd("gateways", "create",
		testFlagFactoryID, "f",
		testFlagFactoryKey, "k",
		testFlagModel, "GW-X",
		testFlagFirmware, "1.0.0",
		testFlagFreq, "1000",
	)
	if err != nil {
		t.Fatalf(fmtUnexpectedError, err)
	}
}

func TestGatewaysCreateMissingModelFails(t *testing.T) {
	err := runCmd("gateways", "create",
		testFlagFactoryID, "f",
		testFlagFactoryKey, "k",
		testFlagFirmware, "1.0.0",
		testFlagFreq, "1000",
	)
	if err == nil {
		t.Fatal("expected error when --model is missing")
	}
}

// -- gateways bulk ------------------------------------------------------------

func TestGatewaysBulkFlagToJSONMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sim/gateways/bulk" {
			t.Errorf(fmtUnexpectedRequest, r.Method, r.URL.Path)
		}
		body := readBody(t, r)
		checkAbsent(t, body, "count")
		checkAbsent(t, body, "factoryId")

		rawFactoryIDs, ok := body["factoryIds"].([]any)
		if !ok {
			t.Fatalf("factoryIds must be an array, got %#v", body["factoryIds"])
		}
		if len(rawFactoryIDs) != 2 {
			t.Fatalf("want 2 factoryIds, got %d", len(rawFactoryIDs))
		}
		if rawFactoryIDs[0] != "fac-bulk-1" || rawFactoryIDs[1] != "fac-bulk-2" {
			t.Fatalf("unexpected factoryIds payload: %#v", rawFactoryIDs)
		}

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
		testFlagFactoryID, "fac-bulk-1",
		testFlagFactoryID, "fac-bulk-2",
		testFlagFactoryKey, "key-bulk",
		testFlagModel, "GW-MINI",
		testFlagFirmware, "1.2.3",
		testFlagFreq, "500",
	)
	if err != nil {
		t.Fatalf("gateways bulk failed: %v", err)
	}
}

// -- gateways get -------------------------------------------------------------

func TestGatewaysGetIntegration(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/sim/gateways/uuid-get-1" {
			t.Errorf(fmtUnexpectedRequest, r.Method, r.URL.Path)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                  "gw-42",
			"managementGatewayId": "uuid-get-1",
			"status":              "online",
			"model":               "GW-X",
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

func TestGatewaysGetNotFound(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	if err := runCmd("gateways", "get", "ghost-uuid"); err == nil {
		t.Error("expected error on 404")
	}
}

// -- sensors add --------------------------------------------------------------

func TestSensorsAddFlagToJSONMapping(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sim/gateways/uuid-gw-1":
			writeJSON(w, http.StatusOK, map[string]any{
				"id":                  "gw-public-5",
				"managementGatewayId": "uuid-gw-1",
			})
		case "/sim/gateways/gw-public-5/sensors":
			if r.Method != http.MethodPost {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			body := readBody(t, r)

			checkKey(t, body, "type", "humidity")
			checkKey(t, body, "minRange", float64(10))
			checkAbsent(t, body, "min")
			checkKey(t, body, "maxRange", float64(90))
			checkAbsent(t, body, "max")
			checkKey(t, body, "algorithm", "uniform_random")

			writeJSON(w, http.StatusCreated, map[string]any{
				"id": "s-uuid", "gatewayId": "gw-public-5",
				"type": "humidity", "minRange": 10, "maxRange": 90, "algorithm": "uniform_random",
			})
		default:
			t.Errorf(fmtUnexpectedPath, r.URL.Path)
		}
	})

	err := runCmd("sensors", "add", "uuid-gw-1",
		"--type", "humidity",
		"--min", "10",
		"--max", "90",
		"--algorithm", "uniform_random",
	)
	if err != nil {
		t.Fatalf("sensors add failed: %v", err)
	}
}

func TestSensorsAddNegativeMinRange(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sim/gateways/uuid-gw-2":
			writeJSON(w, http.StatusOK, map[string]any{
				"id":                  "gw-public-3",
				"managementGatewayId": "uuid-gw-2",
			})
		case "/sim/gateways/gw-public-3/sensors":
			body := readBody(t, r)
			checkKey(t, body, "minRange", float64(-40))
			checkKey(t, body, "maxRange", float64(85))
			writeJSON(w, http.StatusCreated, map[string]any{"id": "s-2", "gatewayId": "gw-public-3"})
		default:
			t.Errorf(fmtUnexpectedPath, r.URL.Path)
		}
	})

	err := runCmd("sensors", "add", "uuid-gw-2",
		"--type", "temperature",
		"--min", "-40",
		"--max", "85",
		"--algorithm", "sine_wave",
	)
	if err != nil {
		t.Fatalf("sensors add with negative min failed: %v", err)
	}
}

// -- anomalies disconnect -----------------------------------------------------

func TestAnomaliesDisconnectDurationFieldName(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sim/gateways/uuid-dc/anomaly/disconnect" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		checkKey(t, body, "duration_seconds", float64(10))
		checkAbsent(t, body, "duration")
		checkAbsent(t, body, "durationSeconds")
		w.WriteHeader(http.StatusNoContent)
	})

	if err := runCmd("anomalies", "disconnect", "uuid-dc", testFlagDuration, "10"); err != nil {
		t.Fatalf("anomalies disconnect failed: %v", err)
	}
}

// -- anomalies network-degradation -------------------------------------------

func TestAnomaliesNetworkDegradationFieldNames(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		checkKey(t, body, "duration_seconds", float64(20))
		checkKey(t, body, "packet_loss_pct", 0.5)
		checkAbsent(t, body, "packetLoss")
		w.WriteHeader(http.StatusNoContent)
	})

	err := runCmd("anomalies", "network-degradation", "uuid-nd",
		testFlagDuration, "20",
		"--packet-loss", "0.5",
	)
	if err != nil {
		t.Fatalf("anomalies network-degradation failed: %v", err)
	}
}

func TestAnomaliesNetworkDegradationPacketLossOmittedWhenNotProvided(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		checkAbsent(t, body, "packet_loss_pct")
		w.WriteHeader(http.StatusNoContent)
	})

	err := runCmd("anomalies", "network-degradation", "uuid-nd", testFlagDuration, "5")
	if err != nil {
		t.Fatalf(fmtUnexpectedError, err)
	}
}

// -- anomalies outlier --------------------------------------------------------

func TestAnomaliesOutlierValueOmittedWhenNotProvided(t *testing.T) {
	newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		checkAbsent(t, body, "value")
		w.WriteHeader(http.StatusNoContent)
	})

	if err := runCmd("anomalies", "outlier", "10"); err != nil {
		t.Fatalf(fmtUnexpectedError, err)
	}
}
