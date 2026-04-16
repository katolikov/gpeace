package resolver

import (
	"fmt"
	"os"
	"strings"

	"github.com/katolikov/gpeace/internal/parser"
)

// Resolution represents the user's choice for a conflict.
type Resolution int

const (
	ResolveCurrent  Resolution = iota // Accept HEAD side
	ResolveIncoming                   // Accept incoming side
	ResolveBoth                       // Accept both (current then incoming)
	ResolveSkip                       // Leave conflict markers in place
)

// Resolve reconstructs the file content from parsed segments and resolutions.
// The resolutions slice must have one entry per conflict segment.
func Resolve(result *parser.ParseResult, resolutions []Resolution) string {
	var buf strings.Builder
	conflictIdx := 0

	for _, seg := range result.Segments {
		if !seg.IsConflict {
			for i, line := range seg.Lines {
				buf.WriteString(line)
				if i < len(seg.Lines)-1 || seg.IsConflict {
					buf.WriteByte('\n')
				}
			}
			// Add newline after the last plain line unless it's the very last segment
			if len(seg.Lines) > 0 {
				buf.WriteByte('\n')
			}
		} else {
			c := seg.Conflict
			r := resolutions[conflictIdx]
			conflictIdx++

			switch r {
			case ResolveCurrent:
				writeLines(&buf, c.CurrentLines)
			case ResolveIncoming:
				writeLines(&buf, c.IncomingLines)
			case ResolveBoth:
				writeLines(&buf, c.CurrentLines)
				writeLines(&buf, c.IncomingLines)
			case ResolveSkip:
				// Preserve original conflict markers
				buf.WriteString(fmt.Sprintf("<<<<<<< %s\n", c.CurrentLabel))
				writeLines(&buf, c.CurrentLines)
				buf.WriteString("=======\n")
				writeLines(&buf, c.IncomingLines)
				buf.WriteString(fmt.Sprintf(">>>>>>> %s\n", c.IncomingLabel))
			}
		}
	}

	return buf.String()
}

func writeLines(buf *strings.Builder, lines []string) {
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
}

// HasUnresolved returns true if any resolution is ResolveSkip.
func HasUnresolved(resolutions []Resolution) bool {
	for _, r := range resolutions {
		if r == ResolveSkip {
			return true
		}
	}
	return false
}

// WriteFile writes content to a file atomically (temp file + rename).
func WriteFile(path, content string) error {
	tmp := path + ".gpeace.tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
