package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/liouk/goot/tui"
)

func main() {
	ctx := context.Background()

	httpClient, err := authenticate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "auth: %s\n", err)
		os.Exit(1)
	}

	client, err := NewClient(ctx, httpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: %s\n", err)
		os.Exit(1)
	}

	var listName string
	if len(os.Args) > 1 {
		listName = os.Args[1]
	}

	appCfg := loadConfig()

	cfg := tui.Config{
		API:              &apiAdapter{client},
		ListName:         listName,
		HiddenListsByID: appCfg.HiddenListsByID,
		HideList:       appCfg.addHiddenList,
	}

	p := tea.NewProgram(tui.New(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %s\n", err)
		os.Exit(1)
	}
}

// apiAdapter bridges the main package Client to the tui.TaskAPI interface.
type apiAdapter struct {
	c *Client
}

func (a *apiAdapter) Lists(ctx context.Context) ([]tui.TaskList, error) {
	lists, err := a.c.Lists(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]tui.TaskList, len(lists))
	for i, l := range lists {
		result[i] = tui.TaskList{ID: l.ID, Title: l.Title}
	}
	return result, nil
}

func (a *apiAdapter) Tasks(ctx context.Context, listID string) ([]tui.Task, error) {
	tasks, err := a.c.Tasks(ctx, listID)
	if err != nil {
		return nil, err
	}
	result := make([]tui.Task, len(tasks))
	for i, t := range tasks {
		result[i] = tui.Task{Title: t.Title, Notes: t.Notes, Due: t.Due}
	}
	return result, nil
}

func (a *apiAdapter) CreateTask(ctx context.Context, listID string, task tui.Task) error {
	return a.c.CreateTask(ctx, listID, Task{
		Title: task.Title,
		Notes: task.Notes,
		Due:   task.Due,
	})
}
