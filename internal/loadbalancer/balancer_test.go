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
            t.Errorf("expected RETRY = %v, got %v", test.want, test.input)
        }
    }
}
