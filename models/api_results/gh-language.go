package api_results

type GHLanguage struct {
	Language  string `json:"language"`
	NumericId int64
}

func ToGHLanguageSlice(langs map[string]int) []GHLanguage {
	values := make([]GHLanguage, 0)
	for k, v := range langs {
		values = append(values, GHLanguage{
			Language:  k,
			NumericId: int64(v),
		})
	}

	return values
}
