package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenPicker screen = iota
	screenCreator
)

// TaskAPI is the interface the TUI uses to talk to Google Tasks.
type TaskAPI interface {
	Lists(ctx context.Context) ([]TaskList, error)
	Tasks(ctx context.Context, listID string) ([]Task, error)
	CreateTask(ctx context.Context, listID string, task Task) error
}

type TaskList struct {
	ID    string
	Title string
}

type Task struct {
	Title string
	Notes string
	Due   string
}

// Config holds the TUI startup configuration.
type Config struct {
	API      TaskAPI
	ListName string // optional: skip picker if this matches a list
}

type model struct {
	cfg        Config
	screen     screen
	picker     pickerModel
	creator    creatorModel
	err        error
	quitting   bool
	submitting bool
	confirming bool // esc was pressed, waiting for y/n
	width      int
}

// listsLoadedMsg is sent when task lists have been fetched.
type listsLoadedMsg struct {
	lists []TaskList
}

// tasksLoadedMsg is sent when tasks for a list have been fetched.
type tasksLoadedMsg struct {
	tasks []Task
}

// taskCreatedMsg is sent when a task has been successfully created.
type taskCreatedMsg struct{}

// successTimeoutMsg is sent after the success message display duration.
type successTimeoutMsg struct{}

// errMsg wraps any error from async operations.
type errMsg struct{ err error }

func New(cfg Config) model {
	return model{
		cfg:    cfg,
		screen: screenPicker,
		picker: newPicker(),
	}
}

func (m model) Init() tea.Cmd {
	return m.loadLists()
}

func (m model) loadLists() tea.Cmd {
	return func() tea.Msg {
		lists, err := m.cfg.API.Lists(context.Background())
		if err != nil {
			return errMsg{err}
		}
		tl := make([]TaskList, len(lists))
		for i, l := range lists {
			tl[i] = TaskList{ID: l.ID, Title: l.Title}
		}
		return listsLoadedMsg{lists: tl}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Error screen: quit on q/esc.
		if m.err != nil {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
			return m, nil
		}

		// Quit confirmation prompt.
		if m.confirming {
			if msg.String() == "y" {
				m.quitting = true
				return m, tea.Quit
			}
			m.confirming = false
			return m, nil
		}

		if msg.String() == "esc" && m.screen == screenCreator {
			m.confirming = true
			return m, nil
		}

		// Go back to picker when backspace is pressed on an empty title field.
		if msg.String() == "backspace" && m.screen == screenCreator &&
			m.creator.focus == fieldTitle && m.creator.inputs[fieldTitle].Value() == "" {
			m.screen = screenPicker
			m.picker.chosen = false
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.picker.list.SetWidth(msg.Width)
		m.creator.width = msg.Width
		return m, nil

	case listsLoadedMsg:
		m.picker = pickerWithLists(m.picker, msg.lists)

		// Skip picker if only one list exists.
		if len(msg.lists) == 1 {
			return m.selectList(msg.lists[0])
		}

		// If a list name was provided, try to skip the picker.
		if m.cfg.ListName != "" {
			for _, l := range msg.lists {
				if matchesListName(l.Title, m.cfg.ListName) {
					return m.selectList(l)
				}
			}
		}
		return m, nil

	case tasksLoadedMsg:
		m.creator = creatorWithTasks(m.creator, msg.tasks)
		return m, nil

	case taskCreatedMsg:
		m.quitting = true
		return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
			return successTimeoutMsg{}
		})

	case successTimeoutMsg:
		return m, tea.Quit

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	switch m.screen {
	case screenPicker:
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(msg)

		if m.picker.chosen {
			return m.selectList(m.picker.selected)
		}
		return m, cmd

	case screenCreator:
		if m.submitting {
			return m, nil
		}

		var cmd tea.Cmd
		m.creator, cmd = m.creator.Update(msg)

		if m.creator.submitted {
			m.submitting = true
			return m, m.createTask()
		}
		return m, cmd
	}

	return m, nil
}

func (m *model) selectList(list TaskList) (tea.Model, tea.Cmd) {
	m.screen = screenCreator
	m.creator = newCreator(list)
	m.creator.width = m.width
	return m, m.loadTasks(list.ID)
}

func (m model) loadTasks(listID string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.cfg.API.Tasks(context.Background(), listID)
		if err != nil {
			return errMsg{err}
		}
		tt := make([]Task, len(tasks))
		for i, t := range tasks {
			tt[i] = Task{Title: t.Title, Notes: t.Notes, Due: t.Due}
		}
		return tasksLoadedMsg{tasks: tt}
	}
}

func (m model) createTask() tea.Cmd {
	task := m.creator.task()
	listID := m.creator.list.ID
	return func() tea.Msg {
		err := m.cfg.API.CreateTask(context.Background(), listID, Task{
			Title: task.Title,
			Notes: task.Notes,
			Due:   task.Due,
		})
		if err != nil {
			return errMsg{err}
		}
		return taskCreatedMsg{}
	}
}

func (m model) View() string {
	if m.err != nil {
		errPrefix := lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3)).Bold(true).Render("Error:")
		return fmt.Sprintf("\n  %s %s\n\n  Press q/esc to quit.\n\n", errPrefix, m.err)
	}
	if m.quitting {
		return "\n  Task created!\n\n"
	}
	if m.confirming {
		return "\n  Quit without saving? (y/n)\n\n"
	}

	switch m.screen {
	case screenPicker:
		return m.picker.View()
	case screenCreator:
		return m.creator.View()
	}
	return ""
}

func matchesListName(title, name string) bool {
	return len(title) > 0 && len(name) > 0 &&
		equalFold(title, name)
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
