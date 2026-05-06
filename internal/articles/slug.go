package articles

import (
	"fmt"
	"strings"

	gslug "github.com/gosimple/slug"
)

func GenerateSlug(title string, exists func(string) bool) string {
	base := gslug.MakeLang(title, "es")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "articulo"
	}
	if len(base) > 80 {
		base = strings.Trim(base[:80], "-")
	}
	s := base
	for i := 2; exists(s); i++ {
		s = fmt.Sprintf("%s-%d", base, i)
	}
	return s
}

