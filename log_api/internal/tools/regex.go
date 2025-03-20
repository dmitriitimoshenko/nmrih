package tools

import "regexp"

var IPRegex = regexp.MustCompile(`\b(?:(25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)\b`)
