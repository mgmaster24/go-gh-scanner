package tokens

type TokenReader interface {
	Fetch(location string) error
	ToTokens() []string
}
