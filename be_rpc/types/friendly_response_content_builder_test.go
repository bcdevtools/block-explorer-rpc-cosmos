package types

import (
	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getDisplayNumber(t *testing.T) {
	tests := []struct {
		num      math.Int
		exponent uint32
		want     string
	}{
		{
			num:      math.NewInt(0),
			exponent: 18,
			want:     "0",
		},
		{
			num:      math.NewInt(20_000_000_000),
			exponent: 9,
			want:     "20",
		},
		{
			num:      math.NewInt(20_000_000_000),
			exponent: 10,
			want:     "2",
		},
		{
			num:      math.NewInt(20_000_000_000),
			exponent: 11,
			want:     "0.2",
		},
		{
			num:      math.NewInt(20_000_000_000),
			exponent: 18,
			want:     "0.00000002",
		},
	}
	for _, tt := range tests {
		t.Run(tt.num.String(), func(t *testing.T) {
			require.Equal(t, tt.want, getDisplayNumber(tt.num, tt.exponent))
		})
	}
}
