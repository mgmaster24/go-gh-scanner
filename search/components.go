package search

import (
	"bufio"
	"path/filepath"
	"regexp"

	"github.com/mgmaster24/go-gh-scanner/utils"
)

var (
	// Angular: selector: 'my-button'  (element selectors only)
	angularSelectorRe = regexp.MustCompile(`selector:\s*['"\x60]([^'"\x60\[\]\s]+)['"\x60]`)
	// React/TS: export const|function|class MyButton
	exportedComponentRe = regexp.MustCompile(`export\s+(?:const|function|class)\s+([A-Z][A-Za-z0-9]+)`)
	// Vue SFC: name: 'MyButton'
	vueNameRe = regexp.MustCompile(`\bname:\s*['"]([^'"]+)['"]`)
)

// ExtractComponentTokens scans source files and returns a deduped list of
// component identifiers per framework:
//   - .ts  → Angular element selectors from @Component decorators
//   - .tsx / .jsx → exported PascalCase names (React components)
//   - .vue → name: value from the component options block
func ExtractComponentTokens(files []string) ([]string, error) {
	seen := make(map[string]bool)
	tokens := make([]string, 0)

	for _, file := range files {
		var re *regexp.Regexp
		switch filepath.Ext(file) {
		case ".ts":
			re = angularSelectorRe
		case ".tsx", ".jsx":
			re = exportedComponentRe
		case ".vue":
			re = vueNameRe
		default:
			continue
		}

		found, err := extractLineMatches(file, re)
		if err != nil {
			return nil, err
		}

		for _, t := range found {
			if !seen[t] {
				seen[t] = true
				tokens = append(tokens, t)
			}
		}
	}

	return tokens, nil
}

func extractLineMatches(filePath string, re *regexp.Regexp) ([]string, error) {
	f, err := utils.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		matches := re.FindAllStringSubmatch(scanner.Text(), -1)
		for _, m := range matches {
			if len(m) > 1 && m[1] != "" {
				results = append(results, m[1])
			}
		}
	}

	return results, scanner.Err()
}
