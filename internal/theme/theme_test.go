package theme

import "testing"

func TestColorsAreDefined(t *testing.T) {
	colors := []struct {
		name  string
		style string
	}{
		{"Background", BgColor},
		{"Primary", PrimaryColor},
		{"Secondary", SecondaryColor},
		{"Success", SuccessColor},
		{"Warning", WarningColor},
		{"Danger", DangerColor},
		{"Muted", MutedColor},
		{"Text", TextColor},
	}
	for _, c := range colors {
		if c.style == "" {
			t.Errorf("color %s is empty", c.name)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 kB"},
		{1048576, "1.0 MB"},
		{1000000000, "1.0 GB"},
	}
	for _, tt := range tests {
		got := FormatSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestSizeBarColor(t *testing.T) {
	c := SizeBarColor(0.1)
	if c != SuccessColor {
		t.Errorf("expected success color for 10%%, got %s", c)
	}
	c = SizeBarColor(0.9)
	if c != DangerColor {
		t.Errorf("expected danger color for 90%%, got %s", c)
	}
}
