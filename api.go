package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/option"
	tasks "google.golang.org/api/tasks/v1"
)

var cetLocation = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	return loc
}()

type TaskList struct {
	ID    string
	Title string
}

type Task struct {
	Title string
	Notes string
	Due   string // YYYY-MM-DD or empty
}

type Client struct {
	svc *tasks.Service
}

func NewClient(ctx context.Context, httpClient *http.Client) (*Client, error) {
	svc, err := tasks.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create tasks service: %w", err)
	}
	return &Client{svc: svc}, nil
}

func (c *Client) Lists(ctx context.Context) ([]TaskList, error) {
	resp, err := c.svc.Tasklists.List().Context(ctx).MaxResults(100).Do()
	if err != nil {
		return nil, fmt.Errorf("list task lists: %w", err)
	}

	lists := make([]TaskList, len(resp.Items))
	for i, item := range resp.Items {
		lists[i] = TaskList{ID: item.Id, Title: item.Title}
	}
	return lists, nil
}

func (c *Client) Tasks(ctx context.Context, listID string) ([]Task, error) {
	resp, err := c.svc.Tasks.List(listID).Context(ctx).MaxResults(100).ShowCompleted(false).Do()
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	result := make([]Task, len(resp.Items))
	for i, item := range resp.Items {
		var due string
		if item.Due != "" {
			if t, err := time.Parse(time.RFC3339, item.Due); err == nil {
				due = t.Format("2006-01-02")
			}
		}
		result[i] = Task{Title: item.Title, Notes: item.Notes, Due: due}
	}
	return result, nil
}

func (c *Client) CreateTask(ctx context.Context, listID string, task Task) error {
	t := &tasks.Task{
		Title: task.Title,
		Notes: task.Notes,
	}

	if task.Due != "" {
		parsed, err := time.ParseInLocation("2006-01-02", task.Due, cetLocation)
		if err != nil {
			return fmt.Errorf("parse due date: %w", err)
		}
		// Use 9am to avoid UTC conversion shifting the date back a day.
		morning := parsed.Add(9 * time.Hour)
		t.Due = morning.Format(time.RFC3339)
	}

	_, err := c.svc.Tasks.Insert(listID, t).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}
