package push

import (
	"strings"
	"testing"
)

func TestValidateAPIURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		// allowed cases
		{
			name:    "valid https",
			url:     "https://api.example.com/v1",
			wantErr: false,
		},
		{
			name:    "http localhost allowed",
			url:     "http://localhost:8000/api/v1",
			wantErr: false,
		},
		{
			name:    "http 127.0.0.1 allowed",
			url:     "http://127.0.0.1:8000/api/v1",
			wantErr: false,
		},

		// https required for non-localhost
		{
			name:    "http non-localhost requires https",
			url:     "http://api.example.com/v1",
			wantErr: true,
			errMsg:  "HTTPS required",
		},
		{
			name:    "http random domain requires https",
			url:     "http://evil.com",
			wantErr: true,
			errMsg:  "HTTPS required",
		},

		// cloud metadata hostnames blocked
		{
			name:    "GCP metadata hostname blocked",
			url:     "https://metadata.google.internal/computeMetadata/v1/",
			wantErr: true,
			errMsg:  "metadata",
		},
		{
			name:    "AWS metadata hostname blocked",
			url:     "https://metadata.amazonaws.com/latest/",
			wantErr: true,
			errMsg:  "metadata",
		},
		{
			name:    "metadata hostname block is case insensitive",
			url:     "https://METADATA.GOOGLE.INTERNAL/foo",
			wantErr: true,
			errMsg:  "metadata",
		},

		// invalid URLs
		{
			name:    "invalid url missing scheme",
			url:     "def-not-a-url",
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
			errMsg:  "empty",
		},
		{
			name:    "missing host",
			url:     "https:///path",
			wantErr: true,
			errMsg:  "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAPIURL(%q) = nil, want error containing %q", tt.url, tt.errMsg)
				} else if tt.errMsg != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
					t.Errorf("ValidateAPIURL(%q) error = %q, want error containing %q", tt.url, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAPIURL(%q) = %v, want nil", tt.url, err)
				}
			}
		})
	}
}
