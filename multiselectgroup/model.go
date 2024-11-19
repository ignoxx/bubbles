package multiselectgroup

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type Option[T comparable] struct {
	ID       T      // required, unique option ID
	Name     string // optional, used to draw option name and still retain the raw value in 'ID'
	Selected bool
}

type Group[T comparable] struct {
	ID      string // required, unique group ID
	Options []Option[T]
}

type MultiSelectGroup[T comparable] struct {
	groups                []Group[T]
	cursor                int
	currentGroup          *Group[T]
	currentGroupIndex     int
	help                  help.Model
	result                *[]Option[T]
	resultHasSelectedOnly bool
	keyMap                KeyMap
}

func NewMultiSelectGroup[T comparable](groups ...Group[T]) *MultiSelectGroup[T] {
	if len(groups) == 0 {
		panic("groups cannot be empty")
	}

	m := MultiSelectGroup[T]{
		groups: groups,
		help:   help.New(),
	}

	m.keyMap = DefaultKeyMap
	m.resultHasSelectedOnly = true

	m.currentGroup = &m.groups[0]
	m.currentGroupIndex = 0
	return &m
}

// writes all selected options into T
func (m *MultiSelectGroup[T]) Value(v *[]Option[T]) *MultiSelectGroup[T] {
	m.result = v
	return m
}

func (m *MultiSelectGroup[T]) KeyMap(k KeyMap) *MultiSelectGroup[T] {
	m.keyMap = k
	return m
}

func (m *MultiSelectGroup[T]) FullResult() *MultiSelectGroup[T] {
	m.resultHasSelectedOnly = false
	return m
}

func (m *MultiSelectGroup[T]) Init() tea.Cmd {
	return nil
}

func (m *MultiSelectGroup[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			} else {
				if m.currentGroupIndex > 0 {
					m.currentGroupIndex -= 1
					m.currentGroup = &m.groups[m.currentGroupIndex]
					m.cursor = len(m.currentGroup.Options) - 1
				}
			}

		case key.Matches(msg, m.keyMap.Down):
			if m.cursor < len(m.currentGroup.Options)-1 {
				m.cursor++
			} else {
				if m.currentGroupIndex < len(m.groups)-1 {
					m.currentGroupIndex += 1
					m.currentGroup = &m.groups[m.currentGroupIndex]
					m.cursor = 0
				}
			}

		case key.Matches(msg, m.keyMap.Toggle):
			if m.cursor >= 0 && m.cursor < len(m.currentGroup.Options) {
				option := &m.currentGroup.Options[m.cursor]
				option.Selected = !option.Selected
			}

		case key.Matches(msg, m.keyMap.Confirm):
			for _, g := range m.groups {
				for _, o := range g.Options {
					if m.resultHasSelectedOnly && !o.Selected {
						continue
					}

					*m.result = append(*m.result, o)
				}
			}

			return m, tea.Quit
		}

	}
	return m, nil
}

func (m *MultiSelectGroup[T]) View() string {
	var (
		sb          strings.Builder
		styles      = huh.ThemeCharm()
		c           = styles.Focused.MultiSelectSelector.Render()
		groupHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		container   = lipgloss.NewStyle().Margin(1)
	)

	for _, g := range m.groups {
		sb.WriteString(groupHeader.Render(g.ID))
		sb.WriteString("\n")

		for i, option := range g.Options {
			if m.currentGroup.ID == g.ID && m.cursor == i {
				sb.WriteString(c)
			} else {
				sb.WriteString(strings.Repeat(" ", lipgloss.Width(c)))
			}

			if option.Selected {
				sb.WriteString(styles.Focused.SelectedPrefix.String())
				sb.WriteString(styles.Focused.SelectedOption.Render(option.Name))
			} else {
				sb.WriteString(styles.Focused.UnselectedPrefix.String())
				sb.WriteString(styles.Focused.UnselectedOption.Render(option.Name))
			}

			sb.WriteString("\n")
		}
	}

	selectionView := sb.String()
	helpView := m.help.View(m.keyMap)
	height := 2 - strings.Count(selectionView, "\n") - strings.Count(helpView, "\n")
	if height <= 0 {
		height = 2
	}

	return container.Render(selectionView + strings.Repeat("\n", height) + helpView)
}
