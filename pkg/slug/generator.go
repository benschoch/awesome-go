package slug

import (
	"github.com/avelino/slugify"
	"strings"
)

func Generate(text string) string {
	// remove slashes to create slugs closer to the GitHub results when parsing markdown
	s := strings.ReplaceAll(text, "/", "")
	return slugify.Slugify(strings.TrimSpace(s))
}
