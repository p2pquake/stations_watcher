package update

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type lineOp int

const (
	opEqual lineOp = iota
	opDelete
	opInsert
)

type lineDiff struct {
	op   lineOp
	line string
}

// lineDiffs computes a line-level diff between a and b using sergi/go-diff.
func lineDiffs(a, b string) []lineDiff {
	dmp := diffmatchpatch.New()
	chars1, chars2, lines := dmp.DiffLinesToChars(a, b)
	diffs := dmp.DiffMain(chars1, chars2, false)
	diffs = dmp.DiffCharsToLines(diffs, lines)

	var out []lineDiff
	for _, d := range diffs {
		// Each Diff.Text contains one or more original lines.
		text := d.Text
		// Drop trailing empty token from final newline split.
		parts := strings.SplitAfter(text, "\n")
		for _, p := range parts {
			if p == "" {
				continue
			}
			// Strip trailing newline; we re-add it during formatting.
			line := strings.TrimSuffix(p, "\n")
			switch d.Type {
			case diffmatchpatch.DiffEqual:
				out = append(out, lineDiff{opEqual, line})
			case diffmatchpatch.DiffDelete:
				out = append(out, lineDiff{opDelete, line})
			case diffmatchpatch.DiffInsert:
				out = append(out, lineDiff{opInsert, line})
			}
		}
	}
	return out
}

// hasChange reports whether the line diff contains any non-equal segments.
func hasChange(lds []lineDiff) bool {
	for _, d := range lds {
		if d.op != opEqual {
			return true
		}
	}
	return false
}

// unifiedDiff renders a unified-diff string with the given context radius.
func unifiedDiff(lds []lineDiff, contextRadius int) string {
	if !hasChange(lds) {
		return ""
	}

	type hunkLine struct {
		op   lineOp
		line string
	}
	type hunk struct {
		oldStart, newStart int
		oldLen, newLen     int
		lines              []hunkLine
	}

	// Track 1-based line numbers as we walk the diff.
	oldLine, newLine := 1, 1
	var hunks []hunk
	var cur *hunk
	// distance since last change, to decide when to close a hunk.
	distSinceChange := 0

	flush := func() {
		if cur != nil {
			hunks = append(hunks, *cur)
			cur = nil
		}
	}

	// First pass: walk through, opening/closing hunks based on context radius.
	// We need leading context too, so we keep a small ring of recent equal lines.
	pending := make([]hunkLine, 0, contextRadius)
	pendingOldStart, pendingNewStart := 0, 0

	for _, d := range lds {
		switch d.op {
		case opEqual:
			if cur == nil {
				// Track recent equal lines for potential leading context.
				if len(pending) == contextRadius {
					pending = pending[1:]
					pendingOldStart++
					pendingNewStart++
				}
				if len(pending) == 0 {
					pendingOldStart = oldLine
					pendingNewStart = newLine
				}
				pending = append(pending, hunkLine{opEqual, d.line})
			} else {
				cur.lines = append(cur.lines, hunkLine{opEqual, d.line})
				cur.oldLen++
				cur.newLen++
				distSinceChange++
				// Close hunk if we've seen enough trailing context AND no further changes are pending.
				// Defer closure check; we close when next change comes too far away.
				if distSinceChange > 2*contextRadius {
					// Trim trailing context beyond contextRadius.
					excess := distSinceChange - contextRadius
					cur.lines = cur.lines[:len(cur.lines)-excess]
					cur.oldLen -= excess
					cur.newLen -= excess
					flush()
					// Start a new pending buffer with the lines we just trimmed,
					// keeping only the last contextRadius for the next hunk.
					pending = pending[:0]
					// We don't have those trimmed lines anymore; just reset based on remaining oldLine/newLine.
					distSinceChange = 0
				}
			}
			oldLine++
			newLine++

		case opDelete, opInsert:
			if cur == nil {
				cur = &hunk{
					oldStart: pendingOldStart,
					newStart: pendingNewStart,
				}
				if len(pending) == 0 {
					cur.oldStart = oldLine
					cur.newStart = newLine
				}
				cur.lines = append(cur.lines, pending...)
				cur.oldLen += len(pending)
				cur.newLen += len(pending)
				pending = pending[:0]
			}
			cur.lines = append(cur.lines, hunkLine{d.op, d.line})
			if d.op == opDelete {
				cur.oldLen++
				oldLine++
			} else {
				cur.newLen++
				newLine++
			}
			distSinceChange = 0
		}
	}
	if cur != nil {
		// Trim trailing context beyond contextRadius.
		if distSinceChange > contextRadius {
			excess := distSinceChange - contextRadius
			cur.lines = cur.lines[:len(cur.lines)-excess]
			cur.oldLen -= excess
			cur.newLen -= excess
		}
		flush()
	}

	var b strings.Builder
	b.WriteString("--- original\n")
	b.WriteString("+++ modified\n")
	for _, h := range hunks {
		oldStart := h.oldStart
		if h.oldLen == 0 {
			oldStart = h.oldStart - 1
		}
		newStart := h.newStart
		if h.newLen == 0 {
			newStart = h.newStart - 1
		}
		fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", oldStart, h.oldLen, newStart, h.newLen)
		for _, ln := range h.lines {
			switch ln.op {
			case opEqual:
				b.WriteString(" ")
			case opDelete:
				b.WriteString("-")
			case opInsert:
				b.WriteString("+")
			}
			b.WriteString(ln.line)
			b.WriteString("\n")
		}
	}
	return b.String()
}
