package tokens

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
)

// Struct that defines common Angular object types.
type NgComponent struct {
	Selector   string `json:"selector"`
	ClassName  string `json:"className"`
	ModuleName string `json:"moduleName"`
}

// Struct that holds Angular tokens.
//
// Implements the TokenReader interface.
type NgTokenReader struct {
	Components []NgComponent
}

// Retrieves a slice of NgComponent from a JSON file.
func (ngCompReader *NgTokenReader) Fetch(fileName string) error {
	ngCompReader.Components = []NgComponent{}
	components, err := reader.ReadJSONData[[]NgComponent](fileName)
	if err != nil {
		return err
	}

	ngCompReader.Components = append(ngCompReader.Components, components...)
	return nil
}

// Turns the slice of NgComponents contained by the reader and turns them
// to a slice of strings
func (ngCompReader *NgTokenReader) ToTokens() []string {
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
