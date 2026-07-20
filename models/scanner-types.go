package models

import (
	"strings"
)

// rangePrefix lists semver range operators and workspace protocol prefixes that
// are stripped from declared versions to yield the base semver string.
var rangePrefix = []string{"workspace:", ">=", "<=", "~^", "^~", "^", "~", ">", "<", "="}

type FragmentStr string

func (f FragmentStr) asString() string {
	return string(f)
}

func (fragment FragmentStr) GetDepVersion(currDep string) (string, bool) {
	fragStr := fragment.asString()
	if !strings.Contains(fragStr, currDep) {
		return "", false
	}

	indexOfDep := strings.Index(fragStr, currDep)
	fragSub := fragStr[indexOfDep:]

	// dependency may be last key in object (no trailing comma)
	endDepIndex := strings.Index(fragSub, ",")
	if endDepIndex == -1 {
		endDepIndex = strings.Index(fragSub, "}")
	}
	if endDepIndex == -1 {
		return "", false
	}

	fragSub = fragSub[:endDepIndex]
	depSections := strings.Split(fragSub, ":")
	if len(depSections) != 2 {
		return "", false
	}

	depVersion := strings.ReplaceAll(depSections[1], "\"", "")
	depVersion = strings.Trim(depVersion, " ")
	depVersion = stripRangePrefix(depVersion)
	return depVersion, depVersion != ""
}

// stripRangePrefix removes semver range operators (^, ~, >=, etc.) and
// workspace protocol prefixes from a declared dependency version string.
// Loops until no further prefix can be stripped (handles "workspace:^" → "").
func stripRangePrefix(v string) string {
	for {
		stripped := v
		for _, prefix := range rangePrefix {
			if strings.HasPrefix(stripped, prefix) {
				stripped = strings.TrimSpace(stripped[len(prefix):])
				break
			}
		}
		if stripped == v {
			return v
		}
		v = stripped
	}
}
