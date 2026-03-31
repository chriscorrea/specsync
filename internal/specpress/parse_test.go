package specpress_test

import (
	"testing"

	"github.com/your-org/specsync/internal/specpress"
)

func TestExtractProjectID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// submitting only a valid UUID
		{
			name:  "valid uuid",
			input: "1f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
			want:  "1f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		},
		{
			name:  "uuid with extra whitespace",
			input: "  2f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c  ",
			want:  "2f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		},

		// submitting valid URL (to extract the project id)
		{
			name:  "spec.press /projects/ URL",
			input: "https://spec.press/projects/7f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
			want:  "7f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		},
		{
			name:  "URL with trailing slash",
			input: "https://spec.press/p/8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c/",
			want:  "8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		},
		{
			name:  "URL with query params",
			input: "https://spec.press/p/9f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c?params=example",
			want:  "9f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
		},

		// invalid inputs
		{
			name:    "non-spec.press URL",
			input:   "https://badsite.com/p/8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
			wantErr: true,
		},
		{
			name:    "internal IP (possible SSRF attempt)",
			input:   "https://169.254.169.254/p/8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c",
			wantErr: true,
		},
		{
			name:    "not a uuid",
			input:   "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Otherwise valid URL missing a uuid",
			input:   "https://spec.press/about",
			wantErr: true,
		},
		{
			name:    "Otherwise valid URL with invalid uuid",
			input:   "https://spec.press/p/5f3a2b1c-not-a-valid-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := specpress.ExtractProjectID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractProjectID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractProjectID() = %v, want %v", got, tt.want)
			}
		})
	}
}
