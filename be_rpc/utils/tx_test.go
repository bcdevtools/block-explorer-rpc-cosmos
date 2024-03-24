package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalizeTransactionHash(t *testing.T) {
	tests := []struct {
		hash  string
		upper bool
		want  string
	}{
		{
			hash:  "0xAA",
			upper: true,
			want:  "0xAA",
		},
		{
			hash:  "0xAA",
			upper: false,
			want:  "0xaa",
		},
		{
			hash:  "0XAA",
			upper: true,
			want:  "0xAA",
		},
		{
			hash:  "0XAA",
			upper: false,
			want:  "0xaa",
		},
		{
			hash:  "AA",
			upper: true,
			want:  "0xAA",
		},
		{
			hash:  "AA",
			upper: false,
			want:  "0xaa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.want, NormalizeTransactionHash(tt.hash, tt.upper))
		})
	}
}
