package intervals

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

type retryPath string

const (
	retryPathReadJSON     retryPath = "read-json"
	retryPathWriteJSON    retryPath = "write-json"
	retryPathNoBodyDelete retryPath = "no-body-delete"
)

var errRetryTransport = errors.New("retry transport")

type currentRetryDecisionCase struct {
	name             string
	path             retryPath
	method           string
	status           int
	err              error
	attempt          int
	retryAfterHeader string
	canceled         bool
	wantRetry        bool
	wantWait         time.Duration
}

var currentRetryDecisionCases = []currentRetryDecisionCase{
	{name: "read get transport error retries", path: retryPathReadJSON, method: http.MethodGet, err: errRetryTransport, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "read get transport error stops when context is canceled", path: retryPathReadJSON, method: http.MethodGet, err: errRetryTransport, attempt: 1, canceled: true, wantRetry: false},
	{name: "read get transport error stops at max attempts", path: retryPathReadJSON, method: http.MethodGet, err: errRetryTransport, attempt: 3, wantRetry: false},
	{name: "read get status 400 does not retry", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusBadRequest, attempt: 1, wantRetry: false},
	{name: "read get status 408 does not retry", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusRequestTimeout, attempt: 1, wantRetry: false},
	{name: "read get status 425 does not retry", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusTooEarly, attempt: 1, wantRetry: false},
	{name: "read get status 429 retries", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusTooManyRequests, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "read get status 429 honors retry-after", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusTooManyRequests, attempt: 1, retryAfterHeader: "1", wantRetry: true, wantWait: time.Second},
	{name: "read get status 429 caps retry-after at max delay", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusTooManyRequests, attempt: 1, retryAfterHeader: "3", wantRetry: true, wantWait: 2 * time.Second},
	{name: "read get status 429 stops when context is canceled", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusTooManyRequests, attempt: 1, canceled: true, wantRetry: false},
	{name: "read get status 500 retries", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusInternalServerError, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "read get status 503 uses exponential delay", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusServiceUnavailable, attempt: 2, wantRetry: true, wantWait: 200 * time.Millisecond},
	{name: "read get status 503 stops at max attempts", path: retryPathReadJSON, method: http.MethodGet, status: http.StatusServiceUnavailable, attempt: 3, wantRetry: false},
	{name: "write post transport error does not retry", path: retryPathWriteJSON, method: http.MethodPost, err: errRetryTransport, attempt: 1, wantRetry: false},
	{name: "write put transport error retries", path: retryPathWriteJSON, method: http.MethodPut, err: errRetryTransport, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write patch transport error retries", path: retryPathWriteJSON, method: http.MethodPatch, err: errRetryTransport, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write delete transport error retries", path: retryPathWriteJSON, method: http.MethodDelete, err: errRetryTransport, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write put transport error stops when context is canceled", path: retryPathWriteJSON, method: http.MethodPut, err: errRetryTransport, attempt: 1, canceled: true, wantRetry: false},
	{name: "write put transport error stops at max attempts", path: retryPathWriteJSON, method: http.MethodPut, err: errRetryTransport, attempt: 3, wantRetry: false},
	{name: "write post status 429 does not retry", path: retryPathWriteJSON, method: http.MethodPost, status: http.StatusTooManyRequests, attempt: 1, wantRetry: false},
	{name: "write put status 408 does not retry", path: retryPathWriteJSON, method: http.MethodPut, status: http.StatusRequestTimeout, attempt: 1, wantRetry: false},
	{name: "write put status 425 does not retry", path: retryPathWriteJSON, method: http.MethodPut, status: http.StatusTooEarly, attempt: 1, wantRetry: false},
	{name: "write put status 429 retries", path: retryPathWriteJSON, method: http.MethodPut, status: http.StatusTooManyRequests, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write patch status 500 retries", path: retryPathWriteJSON, method: http.MethodPatch, status: http.StatusInternalServerError, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write delete status 503 retries", path: retryPathWriteJSON, method: http.MethodDelete, status: http.StatusServiceUnavailable, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write delete status 503 still decides retry when context is canceled", path: retryPathWriteJSON, method: http.MethodDelete, status: http.StatusServiceUnavailable, attempt: 1, canceled: true, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "write put status 503 stops at max attempts", path: retryPathWriteJSON, method: http.MethodPut, status: http.StatusServiceUnavailable, attempt: 3, wantRetry: false},
	{name: "no-body delete transport error retries", path: retryPathNoBodyDelete, method: http.MethodDelete, err: errRetryTransport, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "no-body delete transport error stops when context is canceled", path: retryPathNoBodyDelete, method: http.MethodDelete, err: errRetryTransport, attempt: 1, canceled: true, wantRetry: false},
	{name: "no-body delete transport error stops at max attempts", path: retryPathNoBodyDelete, method: http.MethodDelete, err: errRetryTransport, attempt: 3, wantRetry: false},
	{name: "no-body delete status 408 does not retry", path: retryPathNoBodyDelete, method: http.MethodDelete, status: http.StatusRequestTimeout, attempt: 1, wantRetry: false},
	{name: "no-body delete status 425 does not retry", path: retryPathNoBodyDelete, method: http.MethodDelete, status: http.StatusTooEarly, attempt: 1, wantRetry: false},
	{name: "no-body delete status 429 retries", path: retryPathNoBodyDelete, method: http.MethodDelete, status: http.StatusTooManyRequests, attempt: 1, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "no-body delete status 503 still decides retry when context is canceled", path: retryPathNoBodyDelete, method: http.MethodDelete, status: http.StatusServiceUnavailable, attempt: 1, canceled: true, wantRetry: true, wantWait: 100 * time.Millisecond},
	{name: "no-body delete status 503 stops at max attempts", path: retryPathNoBodyDelete, method: http.MethodDelete, status: http.StatusServiceUnavailable, attempt: 3, wantRetry: false},
}

func TestDecideRetryTruthTable(t *testing.T) {
	t.Parallel()

	client := retryTruthTableClient()
	for _, tc := range currentRetryDecisionCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := retryTruthTableContext(tc.canceled)
			resp := retryTruthTableResponse(tc)
			if resp != nil {
				t.Cleanup(func() { _ = resp.Body.Close() })
			}
			gotRetry, gotWait := client.decideRetry(ctx, tc.method, resp, tc.err, tc.attempt)
			if gotRetry != tc.wantRetry || gotWait != tc.wantWait {
				t.Fatalf("decideRetry() = retry %v wait %v, want retry %v wait %v", gotRetry, gotWait, tc.wantRetry, tc.wantWait)
			}
		})
	}
}

func retryTruthTableResponse(tc currentRetryDecisionCase) *http.Response {
	if tc.err != nil {
		return nil
	}
	resp := &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: http.NoBody}
	if tc.retryAfterHeader != "" {
		resp.Header.Set("Retry-After", tc.retryAfterHeader)
	}
	return resp
}

func retryTruthTableContext(canceled bool) context.Context {
	ctx := context.Background()
	if !canceled {
		return ctx
	}
	canceledCtx, cancel := context.WithCancel(ctx)
	cancel()
	return canceledCtx
}

func retryTruthTableClient() *Client {
	return &Client{retry: RetryConfig{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: 2 * time.Second, Jitter: 0}.WithDefaults()}
}
