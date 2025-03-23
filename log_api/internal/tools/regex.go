package tools

import "regexp"

var (
	IPRegex       = regexp.MustCompile(`\b(?:(25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)\b`)
	DateTimeRegex = regexp.MustCompile(`^(0[1-9]|[12]\d|3[01])\/(0[1-9]|1[0-2])\/\d{4}\s-\s([01]\d|2[0-3]):[0-5]\d:[0-5]\d$`)
)
