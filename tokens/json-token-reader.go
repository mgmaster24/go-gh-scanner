package tokens

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

type JSONTokenReader struct {
	Tokens []string
}

func (jsonRetriever *JSONTokenReader) Fetch(fileName string) error {
	jsonRetriever.Tokens = make([]string, 0)
	tokens, err := reader.ReadJSONData[[]string](fileName)
	if err != nil {
		return err
	}

	jsonRetriever.Tokens = append(jsonRetriever.Tokens, tokens...)
	return nil
}

func (jsonRetriever *JSONTokenReader) ToTokens() []string {
	return utils.RemoveDuplicates(jsonRetriever.Tokens)
}
