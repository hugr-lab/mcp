package auth

import (
	"context"
	"net/http"
)

type tokenCtxKey string

const (
	tokenCtxKeyName tokenCtxKey = "hugr-token"
)

type HugrTransport struct {
	keyHeader string
	c         http.RoundTripper
}

type Option func(*HugrTransport)

func WithSecretHeaderName(name string) Option {
	return func(h *HugrTransport) {
		h.keyHeader = name
	}
}

func WithBaseTransport(c http.RoundTripper) Option {
	return func(h *HugrTransport) {
		h.c = c
	}
}

func New(opts ...Option) *HugrTransport {
	h := &HugrTransport{}
	for _, o := range opts {
		o(h)
	}
	if h.c == nil {
		h.c = http.DefaultTransport
	}
	if h.keyHeader == "" {
		h.keyHeader = "x-hugr-api-key"
	}
	return h
}

func CtxWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenCtxKeyName, token)
}

func CtxWithAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, tokenCtxKeyName, nil)
}

func tokenFromCtx(ctx context.Context) (string, bool) {
	token := ctx.Value(tokenCtxKeyName)
	if token == nil {
		return "", false
	}
	if token, ok := token.(string); ok {
		return token, true
	}
	return "", false
}

func (t *HugrTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if token, ok := tokenFromCtx(req.Context()); ok {
		req.Header.Del(t.keyHeader)
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return t.c.RoundTrip(req)
}
