package paperless

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "date-only format",
			json:     `"2024-01-15"`,
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC3339 format",
			json:     `"2024-01-15T10:30:45Z"`,
			expected: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC3339 with timezone",
			json:     `"2024-01-15T10:30:45-05:00"`,
			expected: time.Date(2024, 1, 15, 15, 30, 45, 0, time.UTC), // UTC equivalent
			wantErr:  false,
		},
		{
			name:     "RFC3339Nano format",
			json:     `"2024-01-15T10:30:45.123456789Z"`,
			expected: time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC),
			wantErr:  false,
		},
		{
			name:    "null value",
			json:    "null",
			wantErr: false,
		},
		{
			name:    "invalid format",
			json:    `"invalid-date"`,
			wantErr: true,
		},
		{
			name:    "empty string",
			json:    `""`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := json.Unmarshal([]byte(tt.json), &d)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.name == "null value" {
				// For null, we don't check the value
				return
			}

			actual := time.Time(d)
			if !actual.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestDate_MarshalJSON(t *testing.T) {
	d := Date(time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC))

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"2024-01-15"`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestDate_Time(t *testing.T) {
	expected := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	d := Date(expected)

	actual := d.Time()
	if !actual.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestDate_String(t *testing.T) {
	d := Date(time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC))

	expected := "2024-01-15"
	actual := d.String()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
