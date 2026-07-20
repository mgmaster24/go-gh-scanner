package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mgmaster24/go-gh-scanner/search"
	"github.com/mgmaster24/go-gh-scanner/utils"
)

func main() {
	repoPath := flag.String("path", "", "path to local design system repo root (required)")
	packageDirs := flag.String("packages", "packages/ng-lib,packages/react-lib,packages/vue-lib", "comma-separated package subdirs to scan")
	out := flag.String("out", "tokens/tokens.json", "output JSON file path")
	flag.Parse()

	if *repoPath == "" {
		fmt.Fprintln(os.Stderr, "usage: gen-tokens -path /path/to/design-system [-packages pkg1,pkg2] [-out tokens.json]")
		os.Exit(1)
	}

	exts := []string{".ts", ".tsx", ".jsx", ".vue"}
	var allFiles []string

	for _, pkg := range strings.Split(*packageDirs, ",") {
		pkg = strings.TrimSpace(pkg)
		// Prefer the src subdirectory to avoid scanning node_modules/dist
		srcDir := filepath.Join(*repoPath, pkg, "src")
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			srcDir = filepath.Join(*repoPath, pkg)
		}
		files, err := utils.GetFilesByExtension(srcDir, exts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error scanning %s: %v\n", srcDir, err)
			os.Exit(1)
		}
		allFiles = append(allFiles, files...)
	}

	tokens, err := search.ExtractComponentTokens(allFiles)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error extracting tokens:", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error marshaling tokens:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*out, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "error writing tokens:", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d tokens to %s\n", len(tokens), *out)
}
