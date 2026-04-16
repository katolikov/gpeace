package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConflictBlock holds the two sides of a single merge conflict.
type ConflictBlock struct {
	StartLine     int      // 0-based line index of <<<<<<< marker
	EndLine       int      // 0-based line index of >>>>>>> marker
	CurrentLines  []string // lines from HEAD side (without trailing \n)
	IncomingLines []string // lines from incoming side (without trailing \n)
	CurrentLabel  string   // text after <<<<<<< (e.g., "HEAD")
	IncomingLabel string   // text after >>>>>>> (e.g., "feature-branch")
}

// Segment represents either plain text or a conflict region in the file.
// Exactly one of Lines or Conflict is meaningful based on IsConflict.
type Segment struct {
	IsConflict bool
	Lines      []string       // plain text lines (when IsConflict=false)
	Conflict   *ConflictBlock // conflict data (when IsConflict=true)
}

// ParseResult holds the complete parse of a single file.
type ParseResult struct {
	FilePath string
	Segments []Segment
	Valid    bool
	Error    string
}

// Conflicts returns just the ConflictBlock pointers from the segments.
func (r *ParseResult) Conflicts() []*ConflictBlock {
	var out []*ConflictBlock
	for i := range r.Segments {
		if r.Segments[i].IsConflict {
			out = append(out, r.Segments[i].Conflict)
		}
	}
	return out
}

type parserState int

const (
	stateNormal   parserState = iota
	stateCurrent              // inside <<<<<<< ... =======
	stateIncoming             // inside ======= ... >>>>>>>
)

const (
	markerCurrent  = "<<<<<<<"
	markerSep      = "======="
	markerIncoming = ">>>>>>>"
)

// ParseFile reads a file and returns its parsed segments.
// Returns a ParseResult with Valid=false if conflict markers are malformed.
func ParseFile(filePath string) *ParseResult {
	f, err := os.Open(filePath)
	if err != nil {
		return &ParseResult{
			FilePath: filePath,
			Valid:    false,
			Error:    fmt.Sprintf("cannot open file: %v", err),
		}
	}
	defer f.Close()

	return ParseLines(filePath, scanLines(f))
}

func scanLines(f *os.File) []string {
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// ParseLines parses conflict markers from pre-split lines.
// Exported for testing without needing actual files.
func ParseLines(filePath string, lines []string) *ParseResult {
	state := stateNormal
	var segments []Segment
	var plainBuf []string
	var currentBlock *ConflictBlock

	flushPlain := func() {
		if len(plainBuf) > 0 {
			segments = append(segments, Segment{
				IsConflict: false,
				Lines:      plainBuf,
			})
			plainBuf = nil
		}
	}

	for i, line := range lines {
		switch state {
		case stateNormal:
			if strings.HasPrefix(line, markerCurrent) && !strings.HasPrefix(line, markerCurrent+"<") {
				flushPlain()
				label := strings.TrimSpace(line[len(markerCurrent):])
				currentBlock = &ConflictBlock{
					StartLine:    i,
					CurrentLabel: label,
				}
				state = stateCurrent
			} else if line == markerSep {
				return &ParseResult{
					FilePath: filePath,
					Valid:    false,
					Error:    fmt.Sprintf("unexpected ======= at line %d outside conflict block", i+1),
				}
			} else if strings.HasPrefix(line, markerIncoming) && !strings.HasPrefix(line, markerIncoming+">") {
				return &ParseResult{
					FilePath: filePath,
					Valid:    false,
					Error:    fmt.Sprintf("unexpected >>>>>>> at line %d outside conflict block", i+1),
				}
			} else {
				plainBuf = append(plainBuf, line)
			}

		case stateCurrent:
			if line == markerSep {
				state = stateIncoming
			} else if strings.HasPrefix(line, markerCurrent) && !strings.HasPrefix(line, markerCurrent+"<") {
				return &ParseResult{
					FilePath: filePath,
					Valid:    false,
					Error:    fmt.Sprintf("nested <<<<<<< at line %d inside conflict block", i+1),
				}
			} else if strings.HasPrefix(line, markerIncoming) && !strings.HasPrefix(line, markerIncoming+">") {
				return &ParseResult{
					FilePath: filePath,
					Valid:    false,
					Error:    fmt.Sprintf("unexpected >>>>>>> at line %d without ======= separator", i+1),
				}
			} else {
				currentBlock.CurrentLines = append(currentBlock.CurrentLines, line)
			}

		case stateIncoming:
			if strings.HasPrefix(line, markerIncoming) && !strings.HasPrefix(line, markerIncoming+">") {
				label := strings.TrimSpace(line[len(markerIncoming):])
				currentBlock.IncomingLabel = label
				currentBlock.EndLine = i
				segments = append(segments, Segment{
					IsConflict: true,
					Conflict:   currentBlock,
				})
				currentBlock = nil
				state = stateNormal
			} else if strings.HasPrefix(line, markerCurrent) && !strings.HasPrefix(line, markerCurrent+"<") {
				return &ParseResult{
					FilePath: filePath,
					Valid:    false,
					Error:    fmt.Sprintf("nested <<<<<<< at line %d inside conflict block", i+1),
				}
			} else {
				currentBlock.IncomingLines = append(currentBlock.IncomingLines, line)
			}
		}
	}

	if state != stateNormal {
		return &ParseResult{
			FilePath: filePath,
			Valid:    false,
			Error:    fmt.Sprintf("unterminated conflict block starting at line %d", currentBlock.StartLine+1),
		}
	}

	// Flush remaining plain text
	flushPlain()

	return &ParseResult{
		FilePath: filePath,
		Segments: segments,
		Valid:    true,
	}
}
