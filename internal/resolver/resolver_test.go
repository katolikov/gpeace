package resolver

import (
	"testing"

	"github.com/katolikov/gpeace/internal/parser"
)

func makeParsed(lines []string) *parser.ParseResult {
	return parser.ParseLines("test.go", lines)
}

func TestResolve_Current(t *testing.T) {
	result := makeParsed([]string{
		"before",
		"<<<<<<< HEAD",
		"current code",
		"=======",
		"incoming code",
		">>>>>>> feature",
		"after",
	})
	out := Resolve(result, []Resolution{ResolveCurrent})
	expected := "before\ncurrent code\nafter\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestResolve_Incoming(t *testing.T) {
	result := makeParsed([]string{
		"before",
		"<<<<<<< HEAD",
		"current code",
		"=======",
		"incoming code",
		">>>>>>> feature",
		"after",
	})
	out := Resolve(result, []Resolution{ResolveIncoming})
	expected := "before\nincoming code\nafter\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestResolve_Both(t *testing.T) {
	result := makeParsed([]string{
		"before",
		"<<<<<<< HEAD",
		"current code",
		"=======",
		"incoming code",
		">>>>>>> feature",
		"after",
	})
	out := Resolve(result, []Resolution{ResolveBoth})
	expected := "before\ncurrent code\nincoming code\nafter\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestResolve_Skip(t *testing.T) {
	result := makeParsed([]string{
		"before",
		"<<<<<<< HEAD",
		"current code",
		"=======",
		"incoming code",
		">>>>>>> feature",
		"after",
	})
	out := Resolve(result, []Resolution{ResolveSkip})
	expected := "before\n<<<<<<< HEAD\ncurrent code\n=======\nincoming code\n>>>>>>> feature\nafter\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestResolve_MultipleConflicts(t *testing.T) {
	result := makeParsed([]string{
		"top",
		"<<<<<<< HEAD",
		"a",
		"=======",
		"b",
		">>>>>>> branch",
		"middle",
		"<<<<<<< HEAD",
		"c",
		"=======",
		"d",
		">>>>>>> branch",
		"bottom",
	})
	out := Resolve(result, []Resolution{ResolveCurrent, ResolveIncoming})
	expected := "top\na\nmiddle\nd\nbottom\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestResolve_EmptySections(t *testing.T) {
	result := makeParsed([]string{
		"<<<<<<< HEAD",
		"=======",
		"incoming",
		">>>>>>> branch",
	})
	out := Resolve(result, []Resolution{ResolveCurrent})
	// Empty current -> nothing emitted for the conflict
	expected := ""
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestResolve_EmptyIncoming(t *testing.T) {
	result := makeParsed([]string{
		"<<<<<<< HEAD",
		"current",
		"=======",
		">>>>>>> branch",
	})
	out := Resolve(result, []Resolution{ResolveIncoming})
	expected := ""
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestHasUnresolved(t *testing.T) {
	if HasUnresolved([]Resolution{ResolveCurrent, ResolveIncoming}) {
		t.Error("expected no unresolved")
	}
	if !HasUnresolved([]Resolution{ResolveCurrent, ResolveSkip}) {
		t.Error("expected unresolved")
	}
}
