// tokens.go - Defines a TokenReader Interface as well as the types of internally
// implemented token readers
package tokens

import "fmt"

// Defined type for TokenReaders
type TokenReaderType int

// Currently implement TokenReader types
const (
	NGTokenReaderType   TokenReaderType = iota
	JSONTokenReaderType TokenReaderType = iota
)

// Interface defining how to retrieve tokens for searching
type TokenReader interface {
	Fetch(location string) error
	ToTokens() []string
}

// Returns a token reader based on the TokenReaderType
func CreateTokenReader(tokenReaderType TokenReaderType) (TokenReader, error) {
	switch tokenReaderType {
	case NGTokenReaderType:
		return &NgTokenReader{}, nil
	case JSONTokenReaderType:
		return &JSONTokenReader{}, nil
	}

	return nil, fmt.Errorf("no token reader found for type %v", tokenReaderType)
}
