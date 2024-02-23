package tokens

import (
	"testing"
)

func TestTokenReader(t *testing.T) {
	var tokenRetriever TokenReader = &NgComponentReader{}
	err := tokenRetriever.Fetch("ng-tokens-test.json")
	if err != nil {
		t.Fatal(err)
	}

	tokens := tokenRetriever.ToTokens()
	if len(tokens) == 0 {
		t.Fatal("Failed to get tokens")
	}

	expectedArray := []string{
		"lib-comp-1",
		"LibComponentModule",
		"lib-comp-2",
		"LibService",
		"lib-comp-3",
		"LibComponent3Module",
		"LibAlertService",
		"LibAlertModule"}

	for i, token := range tokens {
		if expectedArray[i] != token {
			t.Fatal("Failed to verify tokens")
		}
	}
}
