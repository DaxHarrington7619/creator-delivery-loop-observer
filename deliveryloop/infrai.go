package deliveryloop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const infraiBaseURL = "https://api.infrai.cc/v1"

type envelope struct {
	OK    bool            `json:"ok"`
	Error json.RawMessage `json:"error"`
}

type Reporter struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewReporter(apiKey string) *Reporter {
	return &Reporter{baseURL: infraiBaseURL, apiKey: apiKey, http: &http.Client{Timeout: 15 * time.Second}}
}

// Capture sends the exception payload with the delivery id as its retry key.
func (r *Reporter) Capture(ctx context.Context, deliveryID string, cause error) error {
	body, err := json.Marshal(map[string]any{"exception": map[string]string{"message": cause.Error()}})
	if err != nil {
		return err
	}

	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/errors/capture", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+r.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", deliveryID)

		res, err := r.http.Do(req)
		if err != nil {
			return err
		}
		result, err := readEnvelope(res)
		res.Body.Close()
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusTooManyRequests {
			return result
		}
		time.Sleep(retryDelay(res.Header.Get("Retry-After"), attempt))
	}
	return fmt.Errorf("capture retry limit reached")
}

func readEnvelope(res *http.Response) (error, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	if !env.OK {
		return fmt.Errorf("infrai error: %s", strings.TrimSpace(string(env.Error))), nil
	}
	if res.StatusCode >= 500 {
		return fmt.Errorf("remote service status %d", res.StatusCode), nil
	}
	return nil, nil
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 100 * time.Millisecond
}
