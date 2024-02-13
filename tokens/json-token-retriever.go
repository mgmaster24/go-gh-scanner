package tokens

import (
	"github.com/mgmaster24/go-gh-scanner/reader"
)

type JSONTokenRetriever struct {
	Tokens []string
}

func (jsonRetriever *JSONTokenRetriever) Fetch(fileName string) error {
	jsonRetriever.Tokens = make([]string, 0)
	var err error
	jsonRetriever.Tokens, err = reader.ReadConfig[[]string](fileName)
	return err
}

func (jsonRetriever *JSONTokenRetriever) ToTokens() []string {
	return RemoveDuplicates(jsonRetriever.Tokens)
}
