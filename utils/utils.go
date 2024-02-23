package utils

import "strings"

func RemoveDuplicates(tokens []string) []string {
	inMap := make(map[string]bool)
	results := []string{}

	for _, v := range tokens {
		if _, ok := inMap[v]; !ok {
			inMap[v] = true
			results = append(results, v)
		}
	}

	return results
}

func StringsMatch(lhs string, rhs string) bool {
	lhsLen, rhsLen := len(lhs), len(rhs)
	if lhsLen == 0 && rhsLen == 0 {
		return true
	}

	if lhsLen > 1 && lhs[0] == '*' && rhsLen == 0 {
		return false
	}

	if lhsLen != 0 && rhsLen != 0 && lhs[0] == rhs[0] {
		return StringsMatch(lhs[1:lhsLen], rhs[1:rhsLen])
	}

	if lhsLen > 0 && lhs[0] == '*' {
		return StringsMatch(lhs[1:lhsLen], rhs) || StringsMatch(lhs, rhs[1:rhsLen])
	}

	return false
}

func IsStrInStrArray(vals []string, strToCheck string) bool {
	for _, val := range vals {
		if strings.Contains(val, "*") {
			if StringsMatch(val, strToCheck) {
				return true
			}
		} else if val == strToCheck {
			return true
		}
	}

	return false
}
