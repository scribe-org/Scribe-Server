// SPDX-License-Identifier: GPL-3.0-or-later

package handlers

import "fmt"

// normalizeMap ensure all map types are converted to a standard map[string]any format.
func normalizeMap(i any) any {
	switch x := i.(type) {

	case map[string]any:
		for k, v := range x {
			x[k] = normalizeMap(v)
		}
		return x

	case map[any]any:
		m2 := map[string]any{}
		for k, v := range x {
			m2[fmt.Sprint(k)] = normalizeMap(v)
		}
		return m2

	case []any:
		for i, v := range x {
			x[i] = normalizeMap(v)
		}
		return x
	}

	return i
}
