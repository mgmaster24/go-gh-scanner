package tokens

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
)

type NgComponent struct {
	Selector   string `json:"selector"`
	ClassName  string `json:"className"`
	ModuleName string `json:"moduleName"`
}

type NgComponentReader struct {
	Components []NgComponent
}

func (ngCompReader *NgComponentReader) Fetch(fileName string) error {
	ngCompReader.Components = []NgComponent{}
	components, err := reader.ReadJSONData[[]NgComponent](fileName)
	if err != nil {
		return err
	}

	ngCompReader.Components = append(ngCompReader.Components, components...)
	return nil
}

func (ngCompReader *NgComponentReader) ToTokens() []string {
	inMap := make(map[string]bool)
	results := []string{}

	for _, v := range ngCompReader.Components {
		token := v.Selector
		if token == "NONE" {
			token = v.ClassName
		}

		ok := inMap[token]
		if !ok {
			inMap[token] = true
			results = append(results, token)
		}

		token = v.ModuleName
		if token != "NONE" {
			ok = inMap[token]
			if !ok {
				inMap[token] = true
				results = append(results, token)
			}
		}
	}

	return results
}
