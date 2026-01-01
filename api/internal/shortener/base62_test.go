package shortener

import "testing"

func TestEncode(t *testing.T) {
	encoder := NewBase62Encoder(7)

	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0000000"},
		{1, "0000001"},
		{62, "0000010"},
		{100, "000001C"},
	}

	for _, tt := range tests {
		result := encoder.Encode(tt.input)
		if result != tt.expected {
			t.Errorf("Encode(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestDecode(t *testing.T) {
	encoder := NewBase62Encoder(7)

	tests := []struct {
		input    string
		expected uint64
	}{
		{"0", 0},
		{"1", 1},
		{"10", 62},
		{"1C", 100},
		{"000001C", 100},
	}

	for _, tt := range tests {
		result, err := encoder.Decode(tt.input)
		if err != nil {
			t.Errorf("Decode(%s) error: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("Decode(%s) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	encoder := NewBase62Encoder(7)

	testNumbers := []uint64{0, 1, 62, 100, 1000, 1000000}

	for _, num := range testNumbers {
		encoded := encoder.Encode(num)
		decoded, err := encoder.Decode(encoded)

		if err != nil {
			t.Errorf("Decode(%s) error: %v", encoded, err)
		}

		if decoded != num {
			t.Errorf("%d -> %s -> %d, eşleşmedi!", num, encoded, decoded)
		}
	}
}
