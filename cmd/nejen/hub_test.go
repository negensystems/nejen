package main

import (
	"reflect"
	"testing"
)

func TestBreadcrumb(t *testing.T) {
	tests := []struct {
		name  string
		stack []level
		want  string
	}{
		{
			name:  "root alone",
			stack: []level{{title: "NEJEN"}},
			want:  "NEJEN",
		},
		{
			name:  "one level down",
			stack: []level{{title: "NEJEN"}, {title: "Theme"}},
			want:  "NEJEN › THEME",
		},
		{
			// The path the breadcrumb exists for: two levels in, where the old
			// in-place expansion left nothing on screen saying where you were.
			name:  "two levels down",
			stack: []level{{title: "NEJEN"}, {title: "Theme"}, {title: "Wallpaper"}},
			want:  "NEJEN › THEME › WALLPAPER",
		},
		{
			// Labels from `each` are whatever the command printed, so they
			// arrive in arbitrary case and have to be normalised like the rest.
			name:  "mixed case is normalised",
			stack: []level{{title: "NEJEN"}, {title: "Theme"}, {title: "rose-pine"}},
			want:  "NEJEN › THEME › ROSE-PINE",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := breadcrumb(tc.stack); got != tc.want {
				t.Errorf("breadcrumb() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLevelLabels(t *testing.T) {
	items := []HubItem{
		{Label: "Wallpaper", Menu: "wallpaper"},
		{Label: "Pick theme", Each: "nejen theme list", Run: "nejen theme set {}"},
		{Label: "Doctor", Run: "nejen doctor"},
	}

	// Submenus and `each` rows both open another level, so both are marked;
	// leaves are padded to the same width so the labels stay in one column.
	want := []string{
		"▸ Wallpaper",
		"▸ Pick theme",
		"  Doctor",
	}

	if got := levelLabels(items); !reflect.DeepEqual(got, want) {
		t.Errorf("levelLabels() = %q, want %q", got, want)
	}
}

// levelLabels feeds pick(), which maps the chosen row back to an item by
// index, so the two must stay the same length and order.
func TestLevelLabelsIndexAlignment(t *testing.T) {
	items := []HubItem{
		{Label: "A", Menu: "sub"},
		{Label: "B", Run: "true"},
		{Label: "C", Each: "echo x", Run: "echo {}"},
	}

	labels := levelLabels(items)
	if len(labels) != len(items) {
		t.Fatalf("levelLabels() returned %d rows for %d items", len(labels), len(items))
	}
	for i, item := range items {
		if !contains(labels[i], item.Label) {
			t.Errorf("row %d is %q, expected it to carry label %q", i, labels[i], item.Label)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		haystack[len(haystack)-len(needle):] == needle
}
