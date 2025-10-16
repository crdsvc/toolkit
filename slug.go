package toolkit

import (
	"fmt"
	"regexp"
	"strings"
)

func (t *Tools) Slugify(s string) (string, error) {
	if s == "" {
		return s, fmt.Errorf("empty string was sent")
	}

	re := regexp.MustCompile(`[^a-z\d]+`)
	slug := strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")

	if len(slug) == 0 {
		return "", fmt.Errorf("slug is empty")
	}
	return slug, nil
}
