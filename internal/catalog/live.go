package catalog

import (
	"strings"

	"toque/internal/tweaks"
)

func LiveProven() []string {
	var out []string
	for _, e := range Catalog {
		if e.Verdict != VerdictProven {
			continue
		}
		if t, err := tweaks.Get(e.ID); err == nil && t.Kind() == tweaks.Live {
			out = append(out, e.ID)
		}
	}
	return out
}

func Titles(ids []string) []string {
	var out []string
	for _, id := range ids {
		if e, ok := Find(id); ok {
			out = append(out, strings.ToLower(e.Title[:1])+e.Title[1:])
		} else {
			out = append(out, id)
		}
	}
	return out
}
