package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type testStruct struct {
	A string      `json:"a"`
	B *testStruct `json:"b,omitempty"`
}

func TestTryConvertAnyStructToMap(t *testing.T) {

	tests := []struct {
		name    string
		source  any
		wantRes map[string]any
		wantErr bool
	}{
		{
			name:    "passing nil",
			source:  nil,
			wantRes: map[string]any{},
			wantErr: false,
		},
		{
			name:    "passing string",
			source:  "a",
			wantErr: true,
		},
		{
			name:    "passing number",
			source:  int64(1),
			wantErr: true,
		},
		{
			name: "passing struct",
			source: func() any {
				return testStruct{
					A: "a1",
					B: &testStruct{
						A: "a2",
						B: &testStruct{
							A: "a3",
						},
					},
				}
			}(),
			wantRes: map[string]any{
				"a": "a1",
				"b": map[string]any{
					"a": "a2",
					"b": map[string]any{
						"a": "a3",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := TryConvertAnyStructToMap(tt.source)
			if tt.wantErr {
				require.Error(t, err)
				require.NotNil(t, gotRes)
				require.Empty(t, gotRes)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantRes, gotRes)
		})
	}
}

//goland:noinspection GoRedundantConversion
func TestTryGetMapValueAsType(t *testing.T) {
	_map := map[string]any{
		"string": "string",
		"int":    int(1),
		"int8":   int8(6),
		"int16":  int16(16),
		"ts": testStruct{
			A: "a",
		},
	}

	t.Run("angel cases", func(t *testing.T) {
		s, ok := TryGetMapValueAsType[string](_map, "string")
		require.True(t, ok)
		require.Equal(t, "string", s)

		i, ok := TryGetMapValueAsType[int](_map, "int")
		require.True(t, ok)
		require.Equal(t, int(1), i)

		i8, ok := TryGetMapValueAsType[int8](_map, "int8")
		require.True(t, ok)
		require.Equal(t, int8(6), i8)

		i16, ok := TryGetMapValueAsType[int16](_map, "int16")
		require.True(t, ok)
		require.Equal(t, int16(16), i16)

		ts, ok := TryGetMapValueAsType[testStruct](_map, "ts")
		require.True(t, ok)
		require.Equal(t, testStruct{A: "a"}, ts)
	})

	t.Run("mis-match cases", func(t *testing.T) {
		_, ok := TryGetMapValueAsType[string](_map, "int")
		require.False(t, ok)

		_, ok = TryGetMapValueAsType[int](_map, "string")
		require.False(t, ok)

		_, ok = TryGetMapValueAsType[int8](_map, "int16")
		require.False(t, ok)

		_, ok = TryGetMapValueAsType[int16](_map, "int8")
		require.False(t, ok)

		_, ok = TryGetMapValueAsType[testStruct](_map, "int16")
		require.False(t, ok)
	})
}
