package models

import (
	"strings"
)

type FragmentStr string

func (f FragmentStr) AsString() string {
	return string(f)
}

func ToFragmentStr(str string) FragmentStr {
	return FragmentStr(str)
}

func (fragment FragmentStr) GetDepVersion(currDep string) (string, bool) {
	fragStr := fragment.AsString()
	depVersion := ""
	ok := false
	if strings.Contains(fragStr, currDep) {
		indexOfDep := strings.Index(fragStr, currDep)
		fragSub := fragStr[indexOfDep:]
		endDepIndex := strings.Index(fragSub, ",")
		fragSub = fragSub[:endDepIndex]
		depSections := strings.Split(fragSub, ":")
		depLength := len(depSections)
		ok = depLength == 2
		depVersion = strings.ReplaceAll(depSections[depLength-1], "\"", "")
		depVersion = strings.Trim(depVersion, " ")
	}

	return depVersion, ok
}
