// utils.go - Provide generic utility methods.
package utils

import (
	"fmt"
	"strings"
)

// Remove any duplicate tokens for a slice of tokens
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

func RemoveDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		} else {
			fmt.Println("Found Dupe", item)
		}
	}
	return list
}

// Do the strings match
//
// Handles * wildcard strings as well as concrete strings
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

// Does the provided string exist in the provided slice of strings
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

func PanicIfError(e error) {
	if e != nil {
		panic(e)
	}
}
