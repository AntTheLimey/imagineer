/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package auth

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestIsSecureRequest(t *testing.T) {
	tests := []struct {
		name           string
		forwardedProto string
		hasTLS         bool
		want           bool
	}{
		{
			name:           "X-Forwarded-Proto https header present",
			forwardedProto: "https",
			hasTLS:         false,
			want:           true,
		},
		{
			name:           "X-Forwarded-Proto http header present",
			forwardedProto: "http",
			hasTLS:         false,
			want:           false,
		},
		{
			name:           "direct TLS connection without header",
			forwardedProto: "",
			hasTLS:         true,
			want:           true,
		},
		{
			name:           "no header and no TLS",
			forwardedProto: "",
			hasTLS:         false,
			want:           false,
		},
		{
			name:           "X-Forwarded-Proto https with TLS (both present)",
			forwardedProto: "https",
			hasTLS:         true,
			want:           true,
		},
		{
			name:           "X-Forwarded-Proto http with TLS (proxy downgrade)",
			forwardedProto: "http",
			hasTLS:         true,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				Header: make(http.Header),
			}

			if tt.forwardedProto != "" {
				r.Header.Set("X-Forwarded-Proto", tt.forwardedProto)
			}

			if tt.hasTLS {
				r.TLS = &tls.ConnectionState{}
			}

			got := isSecureRequest(r)
			if got != tt.want {
				t.Errorf("isSecureRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
