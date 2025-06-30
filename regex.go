package main

import (
	"net"
	"regexp"
)

func isValidBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	allowedChars := regexp.MustCompile(`^[a-z0-9.-]+$`)
	if !allowedChars.MatchString(name) {
		return false
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}

	if regexp.MustCompile(`[.-]{2,}`).MatchString(name) {
		return false
	}

	if ip := net.ParseIP(name); ip != nil {
		return false
	}

	return true
}
