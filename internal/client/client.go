package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ── Internal constants ────────────────────────────────────────────────────────

const (
	errRequestFailed = "request failed: %w"
	errBuildRequest  = "failed to build request: %w"
	pathGateways     = "/sim/gateways/"
	pathSensors      = "/sim/sensors/"
)

// ── Domain types ─────────────────────────────────────────────────────────────

// Gateway mirrors the GatewayResponse DTO returned by the backend.
// ID is the public gateway identifier (UUID in current API).
type Gateway struct {
	ID                  string `json:"id"`
	ManagementGatewayID string `json:"managementGatewayId,omitempty"`
	FactoryID           string `json:"factoryId"`
	Model               string `json:"model"`
	FirmwareVersion     string `json:"firmwareVersion"`
	Provisioned         bool   `json:"provisioned"`
	SendFrequencyMs     int    `json:"sendFrequencyMs"`
	Status              string `json:"status"`
	TenantID            string `json:"tenantId"`
	CreatedAt           string `json:"createdAt"`
}

// Sensor mirrors the SensorResponse DTO returned by the backend.
type Sensor struct {
	ID        string  `json:"id"`
	GatewayID string  `json:"gatewayId"`
	Type      string  `json:"type"`
	MinRange  float64 `json:"minRange"`
	MaxRange  float64 `json:"maxRange"`
	Algorithm string  `json:"algorithm"`
	CreatedAt string  `json:"createdAt"`
}

// ── Request types ─────────────────────────────────────────────────────────────

// CreateGatewayRequest is the payload for POST /sim/gateways (single).
type CreateGatewayRequest struct {
	FactoryID       string `json:"factoryId"`
	FactoryKey      string `json:"factoryKey"`
	Model           string `json:"model,omitempty"`
	FirmwareVersion string `json:"firmwareVersion,omitempty"`
	SendFrequencyMs int    `json:"sendFrequencyMs,omitempty"`
}

// BulkCreateGatewaysRequest is the payload for POST /sim/gateways/bulk.
type BulkCreateGatewaysRequest struct {
	Count           int    `json:"count"`
	FactoryID       string `json:"factoryId"`
	FactoryKey      string `json:"factoryKey"`
	Model           string `json:"model,omitempty"`
	FirmwareVersion string `json:"firmwareVersion,omitempty"`
	SendFrequencyMs int    `json:"sendFrequencyMs,omitempty"`
}

// BulkCreateResponse is the response for POST /sim/gateways/bulk.
// HTTP 201 means all succeeded; 207 means partial errors.
type BulkCreateResponse struct {
	Gateways []Gateway `json:"gateways"`
	Errors   []string  `json:"errors"`
}

// AddSensorRequest is the payload for POST /sim/gateways/{id}/sensors.
// The gateway ID in the path is the public gateway identifier (UUID string).
type AddSensorRequest struct {
	Type      string  `json:"type"`
	MinRange  float64 `json:"minRange"`
	MaxRange  float64 `json:"maxRange"`
	Algorithm string  `json:"algorithm"`
}

// NetworkDegradationRequest is the payload for POST /sim/gateways/{id}/anomaly/network-degradation.
// PacketLossPct defaults to 0.3 on the backend when omitted or 0.
type NetworkDegradationRequest struct {
	DurationSeconds int     `json:"duration_seconds"`
	PacketLossPct   float64 `json:"packet_loss_pct,omitempty"`
}

// DisconnectRequest is the payload for POST /sim/gateways/{id}/anomaly/disconnect.
// DurationSeconds must be > 0.
type DisconnectRequest struct {
	DurationSeconds int `json:"duration_seconds"`
}

// OutlierRequest is the payload for POST /sim/sensors/{sensorId}/anomaly/outlier.
// Value is optional; the backend applies its own fallback when absent.
type OutlierRequest struct {
	Value *float64 `json:"value,omitempty"`
}

// ── Client ────────────────────────────────────────────────────────────────────

// Client is an HTTP client for the Simulator backend.
type Client struct {
	baseURL    string
	httpClient *http.Client
	ctx        context.Context
}

// defaultTimeout is the maximum time allowed for a single HTTP request.
const defaultTimeout = 30 * time.Second

// New creates a new Client targeting the given baseURL.
func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		ctx:        context.Background(),
	}
}

// WithContext returns a shallow copy of the client bound to ctx.
func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	cc := *c
	cc.ctx = ctx
	return &cc
}

// ── Gateway endpoints ─────────────────────────────────────────────────────────

// CreateGateway calls POST /sim/gateways (single gateway).
func (c *Client) CreateGateway(req CreateGatewayRequest) (*Gateway, error) {
	resp, err := c.post(pathGateways[:len(pathGateways)-1], req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var gw Gateway
	return &gw, json.NewDecoder(resp.Body).Decode(&gw)
}

// BulkCreateGateways calls POST /sim/gateways/bulk.
func (c *Client) BulkCreateGateways(req BulkCreateGatewaysRequest) (*BulkCreateResponse, error) {
	resp, err := c.post(pathGateways+"bulk", req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var result BulkCreateResponse
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

// ListGateways calls GET /sim/gateways.
func (c *Client) ListGateways() ([]Gateway, error) {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, c.baseURL+pathGateways[:len(pathGateways)-1], nil)
	if err != nil {
		return nil, fmt.Errorf(errBuildRequest, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errRequestFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var gateways []Gateway
	return gateways, json.NewDecoder(resp.Body).Decode(&gateways)
}

// GetGateway calls GET /sim/gateways/{id} (UUID).
func (c *Client) GetGateway(id string) (*Gateway, error) {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, c.baseURL+pathGateways+id, nil)
	if err != nil {
		return nil, fmt.Errorf(errBuildRequest, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errRequestFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var gw Gateway
	return &gw, json.NewDecoder(resp.Body).Decode(&gw)
}

// StartGateway calls POST /sim/gateways/{id}/start (UUID, no body).
func (c *Client) StartGateway(id string) error {
	resp, err := c.post(pathGateways+id+"/start", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// StopGateway calls POST /sim/gateways/{id}/stop (UUID, no body).
func (c *Client) StopGateway(id string) error {
	resp, err := c.post(pathGateways+id+"/stop", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// DeleteGateway calls DELETE /sim/gateways/{id} (UUID).
func (c *Client) DeleteGateway(id string) error {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodDelete, c.baseURL+pathGateways+id, nil)
	if err != nil {
		return fmt.Errorf(errBuildRequest, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// ── Sensor endpoints ──────────────────────────────────────────────────────────

// AddSensor calls POST /sim/gateways/{id}/sensors.
// gatewayID is the public gateway identifier (UUID in current API).
func (c *Client) AddSensor(gatewayID string, req AddSensorRequest) (*Sensor, error) {
	path := pathGateways + gatewayID + "/sensors"
	resp, err := c.post(path, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var sensor Sensor
	return &sensor, json.NewDecoder(resp.Body).Decode(&sensor)
}

// ListSensors calls GET /sim/gateways/{id}/sensors.
// gatewayID is the public gateway identifier (UUID in current API).
func (c *Client) ListSensors(gatewayID string) ([]Sensor, error) {
	url := c.baseURL + pathGateways + gatewayID + "/sensors"
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(errBuildRequest, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errRequestFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var sensors []Sensor
	return sensors, json.NewDecoder(resp.Body).Decode(&sensors)
}

// DeleteSensor calls DELETE /sim/sensors/{sensorId}.
// sensorID is the public sensor identifier (UUID in current API).
func (c *Client) DeleteSensor(sensorID string) error {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodDelete, c.baseURL+pathSensors+sensorID, nil)
	if err != nil {
		return fmt.Errorf(errBuildRequest, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(errRequestFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// ── Anomaly endpoints ─────────────────────────────────────────────────────────

// InjectNetworkDegradation calls POST /sim/gateways/{id}/anomaly/network-degradation (UUID).
// If packetLossPct is 0 it is omitted and the backend defaults to 0.3.
func (c *Client) InjectNetworkDegradation(gatewayID string, durationSeconds int, packetLossPct float64) error {
	req := NetworkDegradationRequest{
		DurationSeconds: durationSeconds,
		PacketLossPct:   packetLossPct,
	}
	resp, err := c.post(pathGateways+gatewayID+"/anomaly/network-degradation", req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// Disconnect calls POST /sim/gateways/{id}/anomaly/disconnect (UUID).
// durationSeconds must be > 0.
func (c *Client) Disconnect(gatewayID string, durationSeconds int) error {
	req := DisconnectRequest{DurationSeconds: durationSeconds}
	resp, err := c.post(pathGateways+gatewayID+"/anomaly/disconnect", req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// InjectOutlier calls POST /sim/sensors/{sensorId}/anomaly/outlier.
// sensorID is the public sensor identifier (UUID in current API).
// value is optional (pass nil to omit).
func (c *Client) InjectOutlier(sensorID string, value *float64) error {
	req := OutlierRequest{Value: value}
	resp, err := c.post(pathSensors+sensorID+"/anomaly/outlier", req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return checkStatus(resp)
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *Client) post(path string, body any) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload: %w", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, c.baseURL+path, r)
	if err != nil {
		return nil, fmt.Errorf(errBuildRequest, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errRequestFailed, err)
	}
	return resp, nil
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
}
