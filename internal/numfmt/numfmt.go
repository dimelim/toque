package numfmt

import (
	"strings"

	"toque/internal/i18n"
)

func Localize(s string) string {
	if i18n.Current() == i18n.EN {
		return s
	}
	return strings.Replace(s, ".", ",", 1)
}
