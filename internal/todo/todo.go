package todo

import "time"

type Todo struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTodo(id int, name string, createdAt time.Time) *Todo {
	return &Todo{id, name, createdAt}
}
