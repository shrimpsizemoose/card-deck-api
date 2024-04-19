package main

import "testing"

func TestParseCardCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // expected number of card codes
	}{
		{"Empty string", "", 0},
		{"Single card code", "AS", 1},
		{"Multiple card codes", "AS,KD,5H", 3},
		{"With space", "AS, KD, 5H", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCardCodes(tt.input)
			if len(got) != tt.want {
				t.Errorf("parseCardCodes(%q) = %v, want %v", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestParseCardCodesTrimSpaces(t *testing.T) {
	input := "AS, KD, 5H"
	want := []string{"AS", "KD", "5H"}
	got := parseCardCodes(input)

	if len(got) != len(want) {
		t.Fatalf("Expected slice length %d, got %d", len(want), len(got))
	}

	for i, w := range want {
		if got[i] != w {
			t.Errorf("At index %d, Expected %s, got %s", i, w, got[i])
		}
	}
}
