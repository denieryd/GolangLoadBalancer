package loadbalancer

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetRetryFromContext(t *testing.T) {
    tests := []struct {
        input int
        want  int
    }{
        {1, 1},
        {2, 2},
        {145, 145},
    }
    r := httptest.NewRequest(http.MethodGet, "/mock", nil)

    for _, test := range tests {
        ctx := context.WithValue(r.Context(), RETRY, test.input)
        r = r.WithContext(ctx)

        if retry := GetRetryFromContext(r); retry != test.want {
            t.Errorf("expected RETRY = %v, got %v", test.want, test.input)
        }
    }

    r = r.WithContext(context.Background())

    if retry := GetRetryFromContext(r); retry != 0 {
        t.Errorf("expected RETRY = %v, got %v", 0, retry)
    }
}

func TestGetAttemptsFromContext(t *testing.T) {
    tests := []struct {
        input int
        want  int
    }{
        {1, 1},
        {2, 2},
        {145, 145},
    }
    r := httptest.NewRequest(http.MethodGet, "/mock", nil)

    for _, test := range tests {
        ctx := context.WithValue(r.Context(), ATTEMPTS, test.input)
        r = r.WithContext(ctx)

        if retry := GetAttemptsFromContext(r); retry != test.want {
            t.Errorf("expected ATTEMPTS = %v, got %v", test.want, test.input)
        }
    }

    r = r.WithContext(context.Background())

    if retry := GetAttemptsFromContext(r); retry != 1 {
        t.Errorf("expected ATTEMPTS = %v, got %v", 1, retry)
    }
}
