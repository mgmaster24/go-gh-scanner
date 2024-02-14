package tokens

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

type JSONTokenRetriever struct {
	Tokens []string
}

func (jsonRetriever *JSONTokenRetriever) Fetch(fileName string) error {
	jsonRetriever.Tokens = make([]string, 0)
	tokens, err := reader.ReadJSONData[[]string](fileName)
	if err != nil {
		return err
	}

	jsonRetriever.Tokens = append(jsonRetriever.Tokens, tokens...)
	return nil
}

func (jsonRetriever *JSONTokenRetriever) ToTokens() []string {
	return utils.RemoveDuplicates(jsonRetriever.Tokens)
}
