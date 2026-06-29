package internal

import "time"

type Priority string

const (
	PriorityNone   Priority = ""
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Done      bool       `json:"done"`
	Priority  Priority   `json:"priority,omitempty"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TodoCollection struct {
	Items         []Todo
	Size          int
	MaxLengthItem int
}

func NewTodo(id int, name string, createdAt time.Time) *Todo {
	return &Todo{ID: id, Name: name, CreatedAt: createdAt, UpdatedAt: createdAt}
}
