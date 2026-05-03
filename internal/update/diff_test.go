package update

import (
	"strings"
	"testing"
)

func TestNoChange(t *testing.T) {
	a := "a\nb\nc\n"
	lds := lineDiffs(a, a)
	if hasChange(lds) {
		t.Fatalf("expected no change")
	}
}

func TestUnifiedDiff(t *testing.T) {
	a := "a\nb\nc\nd\ne\n"
	b := "a\nb\nC\nd\ne\n"
	lds := lineDiffs(a, b)
	if !hasChange(lds) {
		t.Fatal("expected change")
	}
	out := unifiedDiff(lds, 1)
	if !strings.Contains(out, "--- original\n+++ modified\n") {
		t.Fatalf("missing headers: %q", out)
	}
	if !strings.Contains(out, "-c") || !strings.Contains(out, "+C") {
		t.Fatalf("missing change markers: %q", out)
	}
	// context_radius=1 -> one leading/trailing equal line.
	if !strings.Contains(out, " b\n-c\n+C\n d\n") {
		t.Fatalf("unexpected hunk content: %q", out)
	}
}

func TestUnifiedDiffEmptyOldFile(t *testing.T) {
	lds := lineDiffs("", "a\nb\n")
	out := unifiedDiff(lds, 1)
	if !strings.Contains(out, "+a\n+b\n") {
		t.Fatalf("expected insertions: %q", out)
	}
}
