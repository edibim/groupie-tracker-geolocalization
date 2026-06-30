package geocode

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"city and country", "mainz-germany", "Mainz, Germany"},
		{"multi-word city", "north_carolina-usa", "North Carolina, USA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.raw)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
