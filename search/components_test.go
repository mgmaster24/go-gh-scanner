package search

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file %s: %v", name, err)
	}
	return path
}

func TestExtractComponentTokens(t *testing.T) {
	dir := t.TempDir()

	angularFile := writeTemp(t, dir, "button.component.ts", `
@Component({
  selector: 'm2s2-button',
  templateUrl: './button.component.html',
})
export class M2s2ButtonComponent {}

@Component({
  selector: 'm2s2-input',
})
export class M2s2InputComponent {}
`)

	reactFile := writeTemp(t, dir, "Card.tsx", `
import React from 'react';

export const M2s2Card = () => <div />;
export function M2s2Badge() { return null; }
export class M2s2Avatar extends React.Component {}
`)

	vueFile := writeTemp(t, dir, "M2s2Alert.vue", `
<script>
export default {
  name: 'M2s2Alert',
  props: ['message'],
}
</script>
`)

	// Non-matching extension should be ignored
	writeTemp(t, dir, "README.md", `selector: 'should-not-match'`)

	files := []string{angularFile, reactFile, vueFile, filepath.Join(dir, "README.md")}
	tokens, err := ExtractComponentTokens(files)
	if err != nil {
		t.Fatalf("ExtractComponentTokens error: %v", err)
	}

	wantTokens := map[string]bool{
		"m2s2-button":      true,
		"m2s2-input":       true,
		"M2s2Card":         true,
		"M2s2Badge":        true,
		"M2s2Avatar":       true,
		"M2s2Alert":        true,
	}

	for _, tok := range tokens {
		if !wantTokens[tok] {
			t.Errorf("unexpected token %q", tok)
		}
		delete(wantTokens, tok)
	}

	for missing := range wantTokens {
		t.Errorf("missing expected token %q", missing)
	}
}

func TestExtractComponentTokens_Deduplication(t *testing.T) {
	dir := t.TempDir()

	// Same selector in two files — should only appear once.
	writeTemp(t, dir, "a.ts", `@Component({ selector: 'm2s2-button' })`)
	writeTemp(t, dir, "b.ts", `@Component({ selector: 'm2s2-button' })`)

	files := []string{
		filepath.Join(dir, "a.ts"),
		filepath.Join(dir, "b.ts"),
	}
	tokens, err := ExtractComponentTokens(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("expected 1 deduplicated token, got %d: %v", len(tokens), tokens)
	}
}
