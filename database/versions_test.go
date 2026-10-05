// SPDX-License-Identifier: GPL-3.0-or-later

package database

import "testing"

func TestLatestVersionDate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		versions map[string]string
		want     string
	}{
		{
			name: "most recent date across data types",
			versions: map[string]string{
				"nouns_last_modified":     "2026-09-30 03:21:30",
				"verbs_last_modified":     "2026-09-28 21:24:20",
				"profanity_last_modified": "2026-10-01 00:52:27",
			},
			want: "2026-10-01",
		},
		{
			name: "RFC3339 values",
			versions: map[string]string{
				"nouns_last_modified": "2026-06-07T08:35:17+01:00",
				"verbs_last_modified": "2026-08-24T18:37:32Z",
			},
			want: "2026-08-24",
		},
		{
			name: "default dates for missing values are ignored",
			versions: map[string]string{
				"nouns_last_modified":        "2026-09-29 15:47:49",
				"prepositions_last_modified": "1970-01-01",
			},
			want: "2026-09-29",
		},
		{
			name:     "only default dates",
			versions: map[string]string{"nouns_last_modified": "1970-01-01"},
			want:     "",
		},
		{
			name:     "values without dates",
			versions: map[string]string{"nouns_last_modified": "", "verbs_last_modified": "unknown"},
			want:     "",
		},
		{
			name:     "no data types",
			versions: map[string]string{},
			want:     "",
		},
	}

	for _, tc := range cases {
		if got := latestVersionDate(tc.versions); got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
