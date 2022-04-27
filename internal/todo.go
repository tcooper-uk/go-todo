package internal

import "time"

type Todo struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoCollection struct {
	Items         []Todo
	Size          int
	MaxLengthItem int
}

func NewTodo(id int, name string, createdAt time.Time) *Todo {
	return &Todo{id, name, createdAt, createdAt}
}
