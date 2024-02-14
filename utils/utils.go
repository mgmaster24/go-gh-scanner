package utils

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
