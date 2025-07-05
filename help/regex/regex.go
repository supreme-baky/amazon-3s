package regex

import (
	"net"
	"regexp"
	"strings"
	"unicode"
)

func IsValidBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	match, _ := regexp.MatchString(`^[a-z0-9.-]+$`, name)
	if !match {
		return false
	}

	if name[0] == '.' || name[0] == '-' || name[len(name)-1] == '.' || name[len(name)-1] == '-' {
		return false
	}

	if strings.Contains(name, "..") || strings.Contains(name, "--") ||
		strings.Contains(name, "-.") || strings.Contains(name, ".-") {
		return false
	}

	if ip := net.ParseIP(name); ip != nil {
		return false
	}
	for _, r := range name {
		if unicode.IsUpper(r) || r > unicode.MaxASCII {
			return false
		}
	}

	pattern := regexp.MustCompile(`^\.\.`)
	if pattern.MatchString(name) {
		return false
	}

	return true
}
