package storage

import "github.com/tcooper-uk/go-todo/internal/todo"

// Represents store of todo items
type TodoStore interface {
	// List all the items.
	// Returns a slice of todo items.
	GetAllItems() []todo.Todo

	// Get a single todo item by it's unique id.
	// Returns a single todo item.
	GetItem(id int) todo.Todo

	// Add a single item.
	// Returns a count of the amount of items added
	// this will be 1 or -1 indicating an error.
	AddItem(value string) int

	// Delete a single item by it's unique id.
	// Returns a count of the amount of items deleted
	// this will be 1 or -1 indicating an error.
	DeleteItem(ids ...int) int

	// Delete all items from the store.
	// Returns a count of the amount of items deleted.
	DeleteAllItems() int
}
