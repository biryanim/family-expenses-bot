package entity

import "testing"

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		name    string
		kopecks int64
		want    string
	}{
		{name: "whole rubles", kopecks: 35000, want: "350"},
		{name: "rubles and kopecks", kopecks: 19990, want: "199.90"},
		{name: "one kopeck", kopecks: 1, want: "0.01"},
		{name: "zero", kopecks: 0, want: "0"},
		{name: "large", kopecks: 120050, want: "1200.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatAmount(tt.kopecks); got != tt.want {
				t.Errorf("FormatAmount(%d) = %q, want %q", tt.kopecks, got, tt.want)
			}
		})
	}
}
