// Package tui provides an interactive terminal UI for browsing image diffs.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/security"
)

// Panel tracks which panel is active.
type Panel int

const (
	PanelFiles Panel = iota
	PanelDetail
)

// Model is the bubbletea model for the TUI.
type Model struct {
	result   *diff.DiffResult
	events   []security.SecurityEvent
	image1   string
	image2   string
	cursor   int
	offset   int
	width    int
	height   int
	panel    Panel
	search   string
	searching bool
	filtered []*diff.DiffEntry
	quit     bool
}

// New creates a new TUI model.
func New(result *diff.DiffResult, events []security.SecurityEvent, image1, image2 string) Model {
	m := Model{
		result: result,
		events: events,
		image1: image1,
		image2: image2,
		panel:  PanelFiles,
	}
	m.filtered = result.Entries
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m.handleSearchInput(msg)
		}
		return m.handleNormalInput(msg)
	}

	return m, nil
}

func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.searching = false
		if msg.String() == "esc" {
			m.search = ""
			m.filtered = m.result.Entries
			m.cursor = 0
			m.offset = 0
		}
	case "backspace":
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.applySearch()
		}
	default:
		if len(msg.String()) == 1 {
			m.search += msg.String()
			m.applySearch()
		}
	}
	return m, nil
}

func (m *Model) applySearch() {
	if m.search == "" {
		m.filtered = m.result.Entries
	} else {
		m.filtered = make([]*diff.DiffEntry, 0)
		for _, e := range m.result.Entries {
			if strings.Contains(strings.ToLower(e.Path), strings.ToLower(m.search)) {
				m.filtered = append(m.filtered, e)
			}
		}
	}
	m.cursor = 0
	m.offset = 0
}

func (m Model) handleNormalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		m.quit = true
		return m, tea.Quit
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}
	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			listHeight := m.listHeight()
			if m.cursor >= m.offset+listHeight {
				m.offset = m.cursor - listHeight + 1
			}
		}
	case key.Matches(msg, keys.PageUp):
		m.cursor -= m.listHeight()
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.offset = m.cursor
	case key.Matches(msg, keys.PageDown):
		m.cursor += m.listHeight()
		if m.cursor >= len(m.filtered) {
			m.cursor = len(m.filtered) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		listHeight := m.listHeight()
		if m.cursor >= m.offset+listHeight {
			m.offset = m.cursor - listHeight + 1
		}
	case key.Matches(msg, keys.Home):
		m.cursor = 0
		m.offset = 0
	case key.Matches(msg, keys.End):
		m.cursor = len(m.filtered) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
		listHeight := m.listHeight()
		m.offset = m.cursor - listHeight + 1
		if m.offset < 0 {
			m.offset = 0
		}
	case key.Matches(msg, keys.Search):
		m.searching = true
		m.search = ""
	case key.Matches(msg, keys.Tab):
		if m.panel == PanelFiles {
			m.panel = PanelDetail
		} else {
			m.panel = PanelFiles
		}
	case key.Matches(msg, keys.Escape):
		m.search = ""
		m.filtered = m.result.Entries
		m.cursor = 0
		m.offset = 0
	}

	return m, nil
}

func (m Model) listHeight() int {
	h := m.height - 6 // header + summary + help + borders
	if h < 1 {
		h = 1
	}
	return h
}

func (m Model) View() string {
	if m.quit {
		return ""
	}
	if m.width == 0 {
		return "Loading..."
	}

	var sb strings.Builder

	// Header
	header := styleHeader.Render(fmt.Sprintf("rift: %s → %s", m.image1, m.image2))
	sb.WriteString(header)
	sb.WriteString("\n")

	// Summary bar
	summary := fmt.Sprintf("  %s added  %s removed  %s modified  %s security events",
		styleAdded.Render(fmt.Sprintf("%d", m.result.Added)),
		styleRemoved.Render(fmt.Sprintf("%d", m.result.Removed)),
		styleModified.Render(fmt.Sprintf("%d", m.result.Modified)),
		styleSecurity.Render(fmt.Sprintf("%d", len(m.events))),
	)
	sb.WriteString(summary)
	sb.WriteString("\n")

	if m.panel == PanelFiles {
		sb.WriteString(m.renderFileList())
	} else {
		sb.WriteString(m.renderDetail())
	}

	// Search bar
	if m.searching {
		sb.WriteString(fmt.Sprintf("\n/%-20s", m.search))
	} else if m.search != "" {
		sb.WriteString(fmt.Sprintf("\n  filter: %s (%d matches)", m.search, len(m.filtered)))
	}

	// Help
	sb.WriteString("\n")
	sb.WriteString(styleHelp.Render("  ↑/k up  ↓/j down  / search  tab panel  q quit"))

	return sb.String()
}

func (m Model) renderFileList() string {
	var sb strings.Builder
	sb.WriteString(styleDim.Render("  ─── Files ───"))
	sb.WriteString("\n")

	listHeight := m.listHeight()
	end := m.offset + listHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	if len(m.filtered) == 0 {
		sb.WriteString("  No matching entries.\n")
		return sb.String()
	}

	for i := m.offset; i < end; i++ {
		entry := m.filtered[i]
		line := formatEntryLine(entry, m.width-4)

		if i == m.cursor {
			sb.WriteString(styleSelected.Render("  "+line))
		} else {
			sb.WriteString("  " + line)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderDetail() string {
	var sb strings.Builder
	sb.WriteString(styleDim.Render("  ─── Detail ───"))
	sb.WriteString("\n")

	if m.cursor >= len(m.filtered) || len(m.filtered) == 0 {
		sb.WriteString("  No entry selected.\n")
		return sb.String()
	}

	entry := m.filtered[m.cursor]
	sb.WriteString(fmt.Sprintf("  Path:  %s\n", entry.Path))
	sb.WriteString(fmt.Sprintf("  Type:  %s\n", styleForType(entry.Type).Render(entry.Type.String())))

	if entry.Before != nil {
		sb.WriteString(fmt.Sprintf("  Before: size=%d mode=%04o uid=%d gid=%d\n",
			entry.Before.Size, entry.Before.Mode, entry.Before.UID, entry.Before.GID))
	}
	if entry.After != nil {
		sb.WriteString(fmt.Sprintf("  After:  size=%d mode=%04o uid=%d gid=%d\n",
			entry.After.Size, entry.After.Mode, entry.After.UID, entry.After.GID))
	}

	if entry.Type == diff.Modified {
		var flags []string
		if entry.ContentChanged {
			flags = append(flags, "content")
		}
		if entry.ModeChanged {
			flags = append(flags, "mode")
		}
		if entry.UIDChanged {
			flags = append(flags, "uid")
		}
		if entry.GIDChanged {
			flags = append(flags, "gid")
		}
		if entry.LinkTargetChanged {
			flags = append(flags, "link")
		}
		if entry.TypeChanged {
			flags = append(flags, "type")
		}
		if len(flags) > 0 {
			sb.WriteString(fmt.Sprintf("  Changes: %s\n", strings.Join(flags, ", ")))
		}
	}

	// Show security events for this path
	for _, ev := range m.events {
		if ev.Path == entry.Path {
			sb.WriteString(fmt.Sprintf("  %s %s (mode: %04o → %04o)\n",
				styleSecurity.Render("SECURITY:"),
				ev.Kind, ev.Before, ev.After))
		}
	}

	return sb.String()
}

func formatEntryLine(entry *diff.DiffEntry, maxWidth int) string {
	var symbol string
	style := styleForType(entry.Type)

	switch entry.Type {
	case diff.Added:
		symbol = "+"
	case diff.Removed:
		symbol = "-"
	case diff.Modified:
		symbol = "~"
	}

	path := entry.Path
	if len(path) > maxWidth-10 {
		path = "..." + path[len(path)-(maxWidth-13):]
	}

	sizeDelta := ""
	if entry.SizeDelta != 0 {
		if entry.SizeDelta > 0 {
			sizeDelta = fmt.Sprintf(" +%d", entry.SizeDelta)
		} else {
			sizeDelta = fmt.Sprintf(" %d", entry.SizeDelta)
		}
	}

	return style.Render(fmt.Sprintf("%s %s%s", symbol, path, sizeDelta))
}

func styleForType(t diff.ChangeType) lipgloss.Style {
	switch t {
	case diff.Added:
		return styleAdded
	case diff.Removed:
		return styleRemoved
	case diff.Modified:
		return styleModified
	default:
		return lipgloss.NewStyle()
	}
}
