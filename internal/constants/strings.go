// SPDX-License-Identifier: GPL-3.0-or-later

// Package constants provides shared, reusable constant values across the application.
package constants

import "unicode"

// IsAlphaNumeric checks if character is alphanumeric.
func IsAlphaNumeric(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char)
}
