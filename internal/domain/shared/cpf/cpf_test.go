package cpf_test

import (
	"strings"
	"testing"

	"github.com/soat13/oficina-serveless/internal/domain/shared/cpf"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantValue string
	}{
		// success
		{"valid_numbers_only", "52998224725", false, "52998224725"},
		{"valid_with_mask", "529.982.247-25", false, "52998224725"},
		{"valid_with_spaces", "  529.982.247-25  ", false, "52998224725"},
		{"valid_with_noise_chars", "abc 529.982.247-25 xyz", false, "52998224725"},
		{"valid_with_tabs_newlines", "\n\t529.982.247-25\t\n", false, "52998224725"},

		// invalid - check digits
		{"invalid_wrong_check_digit", "52998224724", true, ""},
		{"invalid_common_fake", "12345678900", true, ""},

		// invalid - length
		{"invalid_too_short", "123", true, ""},
		{"invalid_too_long", "529982247250", true, ""},
		{"invalid_mask_becomes_10_digits", "529.982.247-2", true, ""},

		// invalid - non-digit characters
		{"invalid_empty", "", true, ""},
		{"invalid_only_punctuation", "...---", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cpf.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got.String() != tt.wantValue {
				t.Fatalf("expected %q, got %q", tt.wantValue, got.String())
			}
		})
	}

	t.Run("invalid_repeated_digits_000_to_999", func(t *testing.T) {
		for d := byte('0'); d <= byte('9'); d++ {
			input := strings.Repeat(string(d), 11)
			_, err := cpf.Parse(input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", input)
			}
		}
	})
}
