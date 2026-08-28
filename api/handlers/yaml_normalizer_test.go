// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"reflect"
	"testing"
)

func TestNormalizeMap(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{
			name: "normalizes nested YAML maps and arrays",
			input: map[any]any{
				"language": "de",
				1: map[any]any{
					"labels": []any{
						map[any]any{"formal": "Sie"},
					},
				},
			},
			want: map[string]any{
				"language": "de",
				"1": map[string]any{
					"labels": []any{
						map[string]any{"formal": "Sie"},
					},
				},
			},
		},
		{
			name: "normalizes values inside string maps",
			input: map[string]any{
				"metadata": map[any]any{"version": 1},
			},
			want: map[string]any{
				"metadata": map[string]any{"version": 1},
			},
		},
		{
			name:  "returns scalar values unchanged",
			input: "unchanged",
			want:  "unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeMap(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeMap() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
