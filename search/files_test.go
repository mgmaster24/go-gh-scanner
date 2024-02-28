package search

import (
	"testing"
)

func TestFindOccurences(t *testing.T) {
	occurences, err := findTokenOccurences("test.html", "test-comp", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(*occurences) == 0 {
		t.Fatal("No occurences found")
	}

	for _, o := range *occurences {
		if o.File != "test.html" {
			t.Fatal("Incorrect file reference")
		}

		if o.LineNumber != 2 && o.LineNumber != 6 {
			t.Fatal("Incorrect line number")
		}

		if o.Token != "test-comp" {
			t.Fatal("Incorrect token")
		}
	}
}
