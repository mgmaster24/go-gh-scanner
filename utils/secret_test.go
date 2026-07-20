package utils

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSecret_ValueRoundTrip(t *testing.T) {
	s := NewSecret("ghp_supersecrettoken")
	if s.Value() != "ghp_supersecrettoken" {
		t.Errorf("Value() = %q, want original token", s.Value())
	}
}

func TestSecret_DoesNotLeakViaFormatting(t *testing.T) {
	s := NewSecret("ghp_supersecrettoken")

	if got := fmt.Sprintf("%v", s); got != "[REDACTED]" {
		t.Errorf("%%v leaked token: %q", got)
	}
	if got := fmt.Sprintf("%s", s); got != "[REDACTED]" {
		t.Errorf("%%s leaked token: %q", got)
	}
	if got := fmt.Sprintf("%#v", s); got != "[REDACTED]" {
		t.Errorf("%%#v leaked token: %q", got)
	}
}

func TestSecret_DoesNotLeakViaJSON(t *testing.T) {
	type payload struct {
		Token Secret `json:"token"`
	}
	p := payload{Token: NewSecret("ghp_supersecrettoken")}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"token":"[REDACTED]"}` {
		t.Errorf("JSON leaked token: %s", b)
	}
}
