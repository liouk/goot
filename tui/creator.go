package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(10)

	taskItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	taskDueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type field int

const (
	fieldTitle field = iota
	fieldNotes
	fieldDue
	fieldCount
)

type creatorModel struct {
	list       TaskList
	tasks      []Task
	inputs     [fieldCount]textinput.Model
	focus      field
	submitted  bool
	dueTouched bool
}

func newCreator(list TaskList) creatorModel {
	title := textinput.New()
	title.Placeholder = "Task title"
	title.Focus()
	title.CharLimit = 256

	notes := textinput.New()
	notes.Placeholder = "Optional notes"
	notes.CharLimit = 1024

	due := textinput.New()
	due.Placeholder = "YYYY-MM-DD"
	due.CharLimit = 10
	due.Placeholder = "YYYY-MM-DD"

	return creatorModel{
		list:   list,
		inputs: [fieldCount]textinput.Model{title, notes, due},
		focus:  fieldTitle,
	}
}

func creatorWithTasks(c creatorModel, tasks []Task) creatorModel {
	c.tasks = tasks
	return c
}

func (m creatorModel) task() Task {
	return Task{
		Title: m.inputs[fieldTitle].Value(),
		Notes: m.inputs[fieldNotes].Value(),
		Due:   m.inputs[fieldDue].Value(),
	}
}

func (m creatorModel) Update(msg tea.Msg) (creatorModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab", "shift+tab":
			if key.String() == "tab" {
				m.focus = (m.focus + 1) % fieldCount
			} else {
				m.focus = (m.focus - 1 + fieldCount) % fieldCount
			}
			if m.focus == fieldDue && !m.dueTouched {
				m.dueTouched = true
				m.inputs[fieldDue].SetValue(time.Now().Format("2006-01-02"))
			}
			return m, m.updateFocus()

		case "enter":
			if m.inputs[fieldTitle].Value() != "" {
				m.submitted = true
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m *creatorModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
	for i := range fieldCount {
		if field(i) == m.focus {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m creatorModel) View() string {
	existing := m.renderExistingTasks()
	form := m.renderForm()

	left := sectionStyle.Width(40).Render(existing)
	right := sectionStyle.Width(44).Render(form)

	return "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n"
}

func (m creatorModel) renderExistingTasks() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(m.list.Title))
	b.WriteString("\n\n")

	if len(m.tasks) == 0 {
		b.WriteString(hintStyle.Render("  No tasks yet"))
	}

	for _, t := range m.tasks {
		line := taskItemStyle.Render("  " + t.Title)
		if t.Due != "" {
			line += taskDueStyle.Render("  " + t.Due)
		}
		b.WriteString(line + "\n")
	}

	return b.String()
}

func (m creatorModel) renderForm() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("New Task"))
	b.WriteString("\n\n")

	labels := [fieldCount]string{"Title", "Notes", "Due"}
	for i := range fieldCount {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render(labels[i]), m.inputs[i].View()))
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  tab: next field | enter: submit | ctrl+c: quit"))

	return b.String()
}
