package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/katolikov/gpeace/internal/git"
	"github.com/katolikov/gpeace/internal/parser"
	"github.com/katolikov/gpeace/internal/resolver"
)

type phase int

const (
	phaseScanning  phase = iota
	phaseResolving
	phaseDone
	phaseError
)

// Menu option indices
const (
	optCurrent  = 0
	optIncoming = 1
	optBoth     = 2
	optSep      = 3 // separator (not selectable)
	optSkip     = 4
	optSkipFile = 5
)

var menuItems = []string{
	"Accept Current Change",
	"Accept Incoming Change",
	"Accept Both Changes",
	"", // separator
	"Skip This Conflict",
	"Skip Entire File",
}

const menuLen = 6

// fileResult tracks the outcome for a single file.
type fileResult struct {
	path          string
	resolved      int
	skipped       int
	skippedFile   bool
	parseError    bool
	errorMsg      string
}

// Model is the bubbletea model for gpeace.
type Model struct {
	phase    phase
	repoRoot string
	err      error

	// File tracking
	files       []string
	fileIdx     int
	parseResult *parser.ParseResult

	// Conflict tracking
	conflicts   []*parser.ConflictBlock
	conflictIdx int
	resolutions []resolver.Resolution

	// UI state
	cursor int
	width  int
	height int

	// Results
	results []fileResult
}

// NewModel creates a new Model for the given repo root.
func NewModel(repoRoot string) Model {
	return Model{
		phase:    phaseScanning,
		repoRoot: repoRoot,
		width:    80,
		height:   24,
		results:  []fileResult{},
	}
}

// Messages
type filesDetectedMsg struct {
	files []string
	err   error
}

type fileParsedMsg struct {
	result *parser.ParseResult
}

type fileWrittenMsg struct {
	path string
	err  error
}

func detectFiles() tea.Msg {
	files, err := git.GetConflictedFiles()
	return filesDetectedMsg{files: files, err: err}
}

func parseFile(repoRoot, relPath string) tea.Cmd {
	return func() tea.Msg {
		absPath := filepath.Join(repoRoot, relPath)
		result := parser.ParseFile(absPath)
		result.FilePath = relPath
		return fileParsedMsg{result: result}
	}
}

func writeAndStage(repoRoot string, result *parser.ParseResult, resolutions []resolver.Resolution) tea.Cmd {
	return func() tea.Msg {
		absPath := filepath.Join(repoRoot, result.FilePath)
		content := resolver.Resolve(result, resolutions)
		if err := resolver.WriteFile(absPath, content); err != nil {
			return fileWrittenMsg{path: result.FilePath, err: err}
		}
		if !resolver.HasUnresolved(resolutions) {
			if err := git.StageFile(absPath); err != nil {
				return fileWrittenMsg{path: result.FilePath, err: err}
			}
		}
		return fileWrittenMsg{path: result.FilePath}
	}
}

// Init starts scanning for conflicted files.
func (m Model) Init() tea.Cmd {
	return detectFiles
}

// Update handles messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case filesDetectedMsg:
		return m.handleFilesDetected(msg)

	case fileParsedMsg:
		return m.handleFileParsed(msg)

	case fileWrittenMsg:
		return m.handleFileWritten(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.phase == phaseDone || m.phase == phaseError {
		if key.Matches(msg, keys.Quit) || key.Matches(msg, keys.Enter) {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.phase != phaseResolving {
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, keys.Up):
		m.cursor--
		if m.cursor == optSep {
			m.cursor--
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case key.Matches(msg, keys.Down):
		m.cursor++
		if m.cursor == optSep {
			m.cursor++
		}
		if m.cursor >= menuLen {
			m.cursor = menuLen - 1
		}

	case key.Matches(msg, keys.Enter):
		return m.handleSelection()

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleSelection() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case optCurrent:
		m.resolutions[m.conflictIdx] = resolver.ResolveCurrent
	case optIncoming:
		m.resolutions[m.conflictIdx] = resolver.ResolveIncoming
	case optBoth:
		m.resolutions[m.conflictIdx] = resolver.ResolveBoth
	case optSkip:
		m.resolutions[m.conflictIdx] = resolver.ResolveSkip
	case optSkipFile:
		m.results = append(m.results, fileResult{
			path:        m.files[m.fileIdx],
			skippedFile: true,
		})
		return m.advanceFile()
	}

	m.cursor = 0

	// Advance to next conflict or finish file
	if m.conflictIdx+1 < len(m.conflicts) {
		m.conflictIdx++
		return m, nil
	}

	// All conflicts resolved for this file -> write it
	return m, writeAndStage(m.repoRoot, m.parseResult, m.resolutions)
}

func (m Model) handleFilesDetected(msg filesDetectedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.phase = phaseError
		m.err = msg.err
		return m, nil
	}
	if len(msg.files) == 0 {
		m.phase = phaseDone
		return m, nil
	}
	m.files = msg.files
	m.fileIdx = 0
	return m, parseFile(m.repoRoot, m.files[0])
}

func (m Model) handleFileParsed(msg fileParsedMsg) (tea.Model, tea.Cmd) {
	if !msg.result.Valid {
		m.results = append(m.results, fileResult{
			path:       m.files[m.fileIdx],
			parseError: true,
			errorMsg:   msg.result.Error,
		})
		return m.advanceFile()
	}

	conflicts := msg.result.Conflicts()
	if len(conflicts) == 0 {
		m.results = append(m.results, fileResult{
			path: m.files[m.fileIdx],
		})
		return m.advanceFile()
	}

	m.parseResult = msg.result
	m.conflicts = conflicts
	m.conflictIdx = 0
	m.resolutions = make([]resolver.Resolution, len(conflicts))
	m.cursor = 0
	m.phase = phaseResolving
	return m, nil
}

func (m Model) handleFileWritten(msg fileWrittenMsg) (tea.Model, tea.Cmd) {
	resolved := 0
	skipped := 0
	for _, r := range m.resolutions {
		if r == resolver.ResolveSkip {
			skipped++
		} else {
			resolved++
		}
	}
	fr := fileResult{
		path:     msg.path,
		resolved: resolved,
		skipped:  skipped,
	}
	if msg.err != nil {
		fr.parseError = true
		fr.errorMsg = msg.err.Error()
	}
	m.results = append(m.results, fr)
	return m.advanceFile()
}

func (m Model) advanceFile() (tea.Model, tea.Cmd) {
	m.fileIdx++
	if m.fileIdx >= len(m.files) {
		m.phase = phaseDone
		return m, nil
	}
	m.phase = phaseScanning
	return m, parseFile(m.repoRoot, m.files[m.fileIdx])
}

// View renders the current state.
func (m Model) View() string {
	switch m.phase {
	case phaseScanning:
		return m.viewScanning()
	case phaseResolving:
		return m.viewResolving()
	case phaseDone:
		return m.viewDone()
	case phaseError:
		return m.viewError()
	}
	return ""
}

func (m Model) viewScanning() string {
	return "\n  Scanning for merge conflicts...\n"
}

func (m Model) viewError() string {
	return fmt.Sprintf("\n  %s %v\n\n  Press q to exit.\n",
		errorStyle.Render("Error:"), m.err)
}

func (m Model) viewResolving() string {
	var b strings.Builder

	conflict := m.conflicts[m.conflictIdx]
	w := m.width
	if w < 40 {
		w = 40
	}

	// Header
	header := headerStyle.Width(w).Render(
		fmt.Sprintf("  gpeace   File %d/%d: %s  |  Conflict %d/%d",
			m.fileIdx+1, len(m.files), m.files[m.fileIdx],
			m.conflictIdx+1, len(m.conflicts)))
	b.WriteString(header)
	b.WriteString("\n\n")

	panelW := w - 4
	if panelW < 30 {
		panelW = 30
	}

	// Current change panel
	currentTitle := currentTitleStyle.Render(
		fmt.Sprintf("Current Change (%s)", labelOrDefault(conflict.CurrentLabel, "ours")))
	currentContent := renderCodeBlock(conflict.CurrentLines)
	currentPanel := currentBorderStyle.Width(panelW).Render(
		currentTitle + "\n" + currentContent)
	b.WriteString("  ")
	b.WriteString(currentPanel)
	b.WriteString("\n\n")

	// Incoming change panel
	incomingTitle := incomingTitleStyle.Render(
		fmt.Sprintf("Incoming Change (%s)", labelOrDefault(conflict.IncomingLabel, "theirs")))
	incomingContent := renderCodeBlock(conflict.IncomingLines)
	incomingPanel := incomingBorderStyle.Width(panelW).Render(
		incomingTitle + "\n" + incomingContent)
	b.WriteString("  ")
	b.WriteString(incomingPanel)
	b.WriteString("\n\n")

	// Menu
	for i, item := range menuItems {
		if i == optSep {
			b.WriteString("    ")
			b.WriteString(separatorStyle.Render(strings.Repeat("─", 24)))
			b.WriteString("\n")
			continue
		}
		if i == m.cursor {
			b.WriteString("    ")
			b.WriteString(selectedStyle.Render("> " + item))
		} else {
			b.WriteString("    ")
			b.WriteString(normalStyle.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("↑/↓ navigate  enter select  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(titleStyle.Render("gpeace - Resolution Complete"))
	b.WriteString("\n\n")

	if len(m.results) == 0 && len(m.files) == 0 {
		b.WriteString("  ")
		b.WriteString(successStyle.Render("No merge conflicts found. Everything is clean!"))
		b.WriteString("\n")
	} else if len(m.results) == 0 {
		b.WriteString("  ")
		b.WriteString(successStyle.Render("No merge conflicts found."))
		b.WriteString("\n")
	} else {
		for _, r := range m.results {
			b.WriteString("  ")
			if r.parseError {
				b.WriteString(errorStyle.Render("!"))
				b.WriteString(fmt.Sprintf(" %-40s ", r.path))
				b.WriteString(errorStyle.Render(r.errorMsg))
			} else if r.skippedFile {
				b.WriteString(skipStyle.Render("-"))
				b.WriteString(fmt.Sprintf(" %-40s ", r.path))
				b.WriteString(skipStyle.Render("skipped entirely"))
			} else if r.skipped > 0 {
				b.WriteString(warningStyle.Render("~"))
				b.WriteString(fmt.Sprintf(" %-40s ", r.path))
				b.WriteString(warningStyle.Render(
					fmt.Sprintf("%d resolved, %d skipped (not staged)", r.resolved, r.skipped)))
			} else if r.resolved > 0 {
				b.WriteString(successStyle.Render("✓"))
				b.WriteString(fmt.Sprintf(" %-40s ", r.path))
				b.WriteString(successStyle.Render(
					fmt.Sprintf("%d conflict(s) resolved", r.resolved)))
			} else {
				b.WriteString(skipStyle.Render("-"))
				b.WriteString(fmt.Sprintf(" %-40s ", r.path))
				b.WriteString(skipStyle.Render("no conflicts"))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("Press q or enter to exit."))
	b.WriteString("\n")

	return b.String()
}

func labelOrDefault(label, def string) string {
	if label == "" {
		return def
	}
	return label
}

func renderCodeBlock(lines []string) string {
	if len(lines) == 0 {
		return lipgloss.NewStyle().Foreground(dim).Render("  (empty)")
	}
	var b strings.Builder
	for i, line := range lines {
		num := lineNumStyle.Render(fmt.Sprintf("%d", i+1))
		b.WriteString(num)
		b.WriteString(" │ ")
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
