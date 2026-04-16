package parser

import (
	"strings"
	"testing"
)

func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func TestParseLines_NoConflicts(t *testing.T) {
	input := lines("line 1\nline 2\nline 3")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result.Segments))
	}
	if result.Segments[0].IsConflict {
		t.Fatal("expected plain segment")
	}
	if len(result.Segments[0].Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result.Segments[0].Lines))
	}
	conflicts := result.Conflicts()
	if len(conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestParseLines_SingleConflict(t *testing.T) {
	input := lines("before\n<<<<<<< HEAD\ncurrent code\n=======\nincoming code\n>>>>>>> feature\nafter")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}

	conflicts := result.Conflicts()
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}

	c := conflicts[0]
	if c.CurrentLabel != "HEAD" {
		t.Errorf("expected label HEAD, got %q", c.CurrentLabel)
	}
	if c.IncomingLabel != "feature" {
		t.Errorf("expected label feature, got %q", c.IncomingLabel)
	}
	if len(c.CurrentLines) != 1 || c.CurrentLines[0] != "current code" {
		t.Errorf("unexpected current lines: %v", c.CurrentLines)
	}
	if len(c.IncomingLines) != 1 || c.IncomingLines[0] != "incoming code" {
		t.Errorf("unexpected incoming lines: %v", c.IncomingLines)
	}

	// Check surrounding segments
	if len(result.Segments) != 3 {
		t.Fatalf("expected 3 segments (plain, conflict, plain), got %d", len(result.Segments))
	}
	if result.Segments[0].Lines[0] != "before" {
		t.Errorf("expected 'before', got %q", result.Segments[0].Lines[0])
	}
	if result.Segments[2].Lines[0] != "after" {
		t.Errorf("expected 'after', got %q", result.Segments[2].Lines[0])
	}
}

func TestParseLines_MultipleConflicts(t *testing.T) {
	input := lines("top\n<<<<<<< HEAD\na\n=======\nb\n>>>>>>> branch\nmiddle\n<<<<<<< HEAD\nc\n=======\nd\n>>>>>>> branch\nbottom")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}

	conflicts := result.Conflicts()
	if len(conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d", len(conflicts))
	}
	if conflicts[0].CurrentLines[0] != "a" {
		t.Errorf("first conflict current: %v", conflicts[0].CurrentLines)
	}
	if conflicts[1].CurrentLines[0] != "c" {
		t.Errorf("second conflict current: %v", conflicts[1].CurrentLines)
	}

	// 5 segments: plain, conflict, plain, conflict, plain
	if len(result.Segments) != 5 {
		t.Fatalf("expected 5 segments, got %d", len(result.Segments))
	}
}

func TestParseLines_AdjacentConflicts(t *testing.T) {
	input := lines("<<<<<<< HEAD\na\n=======\nb\n>>>>>>> b1\n<<<<<<< HEAD\nc\n=======\nd\n>>>>>>> b2")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}

	conflicts := result.Conflicts()
	if len(conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d", len(conflicts))
	}
}

func TestParseLines_ConflictAtStartOfFile(t *testing.T) {
	input := lines("<<<<<<< HEAD\nfirst\n=======\nsecond\n>>>>>>> branch\nrest of file")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	if result.Segments[0].IsConflict != true {
		t.Error("expected first segment to be a conflict")
	}
}

func TestParseLines_ConflictAtEndOfFile(t *testing.T) {
	input := lines("start of file\n<<<<<<< HEAD\nlast\n=======\nother\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	lastSeg := result.Segments[len(result.Segments)-1]
	if !lastSeg.IsConflict {
		t.Error("expected last segment to be a conflict")
	}
}

func TestParseLines_EmptyCurrentSection(t *testing.T) {
	input := lines("<<<<<<< HEAD\n=======\nincoming\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	c := result.Conflicts()[0]
	if len(c.CurrentLines) != 0 {
		t.Errorf("expected empty current lines, got %v", c.CurrentLines)
	}
	if len(c.IncomingLines) != 1 {
		t.Errorf("expected 1 incoming line, got %d", len(c.IncomingLines))
	}
}

func TestParseLines_EmptyIncomingSection(t *testing.T) {
	input := lines("<<<<<<< HEAD\ncurrent\n=======\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	c := result.Conflicts()[0]
	if len(c.CurrentLines) != 1 {
		t.Errorf("expected 1 current line, got %d", len(c.CurrentLines))
	}
	if len(c.IncomingLines) != 0 {
		t.Errorf("expected empty incoming lines, got %v", c.IncomingLines)
	}
}

func TestParseLines_BothSectionsEmpty(t *testing.T) {
	input := lines("<<<<<<< HEAD\n=======\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	c := result.Conflicts()[0]
	if len(c.CurrentLines) != 0 || len(c.IncomingLines) != 0 {
		t.Errorf("expected both sections empty, got current=%v incoming=%v", c.CurrentLines, c.IncomingLines)
	}
}

func TestParseLines_EightEqualsNotMarker(t *testing.T) {
	input := lines("========\n=========\n========================================")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("expected 1 plain segment, got %d", len(result.Segments))
	}
	if len(result.Segments[0].Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result.Segments[0].Lines))
	}
}

func TestParseLines_Malformed_OrphanedSeparator(t *testing.T) {
	input := lines("some code\n=======\nmore code")
	result := ParseLines("test.go", input)

	if result.Valid {
		t.Fatal("expected invalid result for orphaned =======")
	}
	if !strings.Contains(result.Error, "unexpected =======") {
		t.Errorf("unexpected error message: %s", result.Error)
	}
}

func TestParseLines_Malformed_OrphanedCloser(t *testing.T) {
	input := lines("some code\n>>>>>>> branch\nmore code")
	result := ParseLines("test.go", input)

	if result.Valid {
		t.Fatal("expected invalid result for orphaned >>>>>>>")
	}
}

func TestParseLines_Malformed_Unterminated(t *testing.T) {
	input := lines("<<<<<<< HEAD\ncurrent\n=======\nincoming")
	result := ParseLines("test.go", input)

	if result.Valid {
		t.Fatal("expected invalid result for unterminated conflict")
	}
	if !strings.Contains(result.Error, "unterminated") {
		t.Errorf("unexpected error message: %s", result.Error)
	}
}

func TestParseLines_Malformed_NestedOpener(t *testing.T) {
	input := lines("<<<<<<< HEAD\n<<<<<<< nested\n=======\nincoming\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if result.Valid {
		t.Fatal("expected invalid result for nested <<<<<<<")
	}
}

func TestParseLines_Malformed_MissingSeparator(t *testing.T) {
	input := lines("<<<<<<< HEAD\ncurrent\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if result.Valid {
		t.Fatal("expected invalid result for missing =======")
	}
}

func TestParseLines_Labels(t *testing.T) {
	input := lines("<<<<<<< HEAD\ncode\n=======\ncode\n>>>>>>> refs/heads/feature-branch")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	c := result.Conflicts()[0]
	if c.CurrentLabel != "HEAD" {
		t.Errorf("expected HEAD, got %q", c.CurrentLabel)
	}
	if c.IncomingLabel != "refs/heads/feature-branch" {
		t.Errorf("expected refs/heads/feature-branch, got %q", c.IncomingLabel)
	}
}

func TestParseLines_MultiLineConflict(t *testing.T) {
	input := lines("<<<<<<< HEAD\nline 1\nline 2\nline 3\n=======\nalt 1\nalt 2\n>>>>>>> branch")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	c := result.Conflicts()[0]
	if len(c.CurrentLines) != 3 {
		t.Errorf("expected 3 current lines, got %d", len(c.CurrentLines))
	}
	if len(c.IncomingLines) != 2 {
		t.Errorf("expected 2 incoming lines, got %d", len(c.IncomingLines))
	}
}

func TestParseLines_EmptyInput(t *testing.T) {
	result := ParseLines("empty.go", nil)

	if !result.Valid {
		t.Fatalf("expected valid for empty input, got error: %s", result.Error)
	}
	if len(result.Segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(result.Segments))
	}
}

func TestParseLines_LineIndices(t *testing.T) {
	input := lines("line0\nline1\n<<<<<<< HEAD\nline3\n=======\nline5\n>>>>>>> branch\nline7")
	result := ParseLines("test.go", input)

	if !result.Valid {
		t.Fatalf("expected valid, got error: %s", result.Error)
	}
	c := result.Conflicts()[0]
	if c.StartLine != 2 {
		t.Errorf("expected StartLine=2, got %d", c.StartLine)
	}
	if c.EndLine != 6 {
		t.Errorf("expected EndLine=6, got %d", c.EndLine)
	}
}
