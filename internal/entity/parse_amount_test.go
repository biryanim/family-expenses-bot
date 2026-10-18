package entity

import (
	"errors"
	"testing"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "whole rubles", input: "350", want: 35000},
		{name: "one fraction digit", input: "199.9", want: 19990},
		{name: "one kopeck", input: "0.01", want: 1},
		{name: "two fraction digits", input: "1200.50", want: 120050},
		{name: "zero fraction", input: "350.0", want: 35000},
		{name: "zero", input: "0", wantErr: true},
		{name: "zero with fraction", input: "0.00", wantErr: true},
		{name: "negative", input: "-5", wantErr: true},
		{name: "plus sign", input: "+5", wantErr: true},
		{name: "three fraction digits", input: "1.005", wantErr: true},
		{name: "exponent", input: "1e3", wantErr: true},
		{name: "letters", input: "abc", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "no integer part", input: ".5", wantErr: true},
		{name: "trailing dot", input: "350.", wantErr: true},
		{name: "comma", input: "350,50", wantErr: true},
		{name: "overflow", input: "99999999999999999999", wantErr: true},
		{name: "overflow after multiply", input: "92233720368547759", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidAmount) {
					t.Errorf("ParseAmount(%q) error = %v, want %v", tt.input, err, ErrInvalidAmount)
				}
				if got != 0 {
					t.Errorf("ParseAmount(%q) = %d, want 0", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseAmount(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseAmount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
