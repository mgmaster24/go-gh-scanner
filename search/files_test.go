package search

import (
	"fmt"
	"testing"
)

func TestFindOccurences(t *testing.T) {
	occurences, err := FindTokenOccurences("test.html", "test-comp")
	if err != nil {
		t.Fatal(err)
	}

	if len(*occurences) == 0 {
		t.Fatal("No occurences found")
	}

	for _, o := range *occurences {
		fmt.Println(o)
	}
}
