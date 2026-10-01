// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNormalizeMap tests recursive conversion of map[any]any into map[string]any.
func TestNormalizeMap(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{
			name:  "nil passthrough",
			input: nil,
			want:  nil,
		},
		{
			name:  "string passthrough",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "int passthrough",
			input: 42,
			want:  42,
		},
		{
			name:  "bool passthrough",
			input: true,
			want:  true,
		},
		{
			name:  "float64 passthrough",
			input: 3.14,
			want:  3.14,
		},
		{
			name:  "empty map[string]any",
			input: map[string]any{},
			want:  map[string]any{},
		},
		{
			name:  "flat map[string]any",
			input: map[string]any{"a": 1, "b": "two"},
			want:  map[string]any{"a": 1, "b": "two"},
		},
		{
			name:  "empty map[any]any",
			input: map[any]any{},
			want:  map[string]any{},
		},
		{
			name: "flat map[any]any with string keys",
			input: map[any]any{
				"x": 10,
				"y": "hello",
			},
			want: map[string]any{
				"x": 10,
				"y": "hello",
			},
		},
		{
			name: "map[any]any with non-string key (int)",
			input: map[any]any{
				42: "answer",
			},
			want: map[string]any{
				"42": "answer",
			},
		},
		{
			name: "map[any]any with bool key",
			input: map[any]any{
				true: "yes",
			},
			want: map[string]any{
				"true": "yes",
			},
		},
		{
			name:  "empty slice",
			input: []any{},
			want:  []any{},
		},
		{
			name:  "slice of scalars",
			input: []any{1, "two", false},
			want:  []any{1, "two", false},
		},
		{
			name: "slice containing map[any]any",
			input: []any{
				map[any]any{"k": "v"},
			},
			want: []any{
				map[string]any{"k": "v"},
			},
		},
		{
			name: "nested map[any]any two levels deep",
			input: map[any]any{
				"outer": map[any]any{
					"inner": "value",
				},
			},
			want: map[string]any{
				"outer": map[string]any{
					"inner": "value",
				},
			},
		},
		{
			name: "map[string]any with map[any]any value",
			input: map[string]any{
				"top": map[any]any{
					"nested": 99,
				},
			},
			want: map[string]any{
				"top": map[string]any{
					"nested": 99,
				},
			},
		},
		{
			name: "map[any]any value is a slice containing map[any]any",
			input: map[any]any{
				"items": []any{
					map[any]any{"id": 1},
					map[any]any{"id": 2},
				},
			},
			want: map[string]any{
				"items": []any{
					map[string]any{"id": 1},
					map[string]any{"id": 2},
				},
			},
		},
		{
			name: "three levels of map[any]any nesting",
			input: map[any]any{
				"l1": map[any]any{
					"l2": map[any]any{
						"l3": "deep",
					},
				},
			},
			want: map[string]any{
				"l1": map[string]any{
					"l2": map[string]any{
						"l3": "deep",
					},
				},
			},
		},
		{
			name: "slice of slices each containing a map[any]any",
			input: []any{
				[]any{map[any]any{"a": 1}},
				[]any{map[any]any{"b": 2}},
			},
			want: []any{
				[]any{map[string]any{"a": 1}},
				[]any{map[string]any{"b": 2}},
			},
		},
		{
			name: "slice with mixed map types",
			input: []any{
				map[string]any{"already": "normalised"},
				map[any]any{"needs": "conversion"},
			},
			want: []any{
				map[string]any{"already": "normalised"},
				map[string]any{"needs": "conversion"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMap(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

