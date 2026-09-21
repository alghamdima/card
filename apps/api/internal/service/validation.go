package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidationError carries a client-safe message and matches the sentinel
// error of the domain it belongs to (ErrInvalidCampaignInput, ErrInvalidCardInput).
type ValidationError struct {
	Message string
	kind    error
}

func (e *ValidationError) Error() string { return e.Message }

func (e *ValidationError) Is(target error) bool { return target == e.kind }

func invalidf(kind error, format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...), kind: kind}
}

var (
	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	slugPattern     = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$`)
	fieldIDPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	imagePattern    = regexp.MustCompile(`^data:image/(?:png|jpeg|jpg|webp|gif);base64,`)
)

func isHexColor(s string) bool { return hexColorPattern.MatchString(s) }

func isImageDataURL(s string) bool { return imagePattern.MatchString(s) }

// cleanText trims the input, drops control characters (keeping newlines and
// tabs) and enforces a maximum length in runes.
func cleanText(s string, maxRunes int) (string, bool) {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	return s, utf8.RuneCountInString(s) <= maxRunes
}
