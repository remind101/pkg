package aws

import (
	"strings"
)

const carrierPrefix = "ot-carrier-"

func addPrefix(s string) string {
	return carrierPrefix + s
}

func removePrefix(s string) string {
	return strings.TrimPrefix(s, carrierPrefix)
}

func hasPrefix(s string) bool {
	return strings.HasPrefix(s, carrierPrefix)
}
