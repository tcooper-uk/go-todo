package storage

import (
	"errors"
	"fmt"
	"os"

	"github.com/tcooper-uk/go-todo/internal"
)

const (
	DB_FILE   = "todo.db"
	JSON_FILE = "todo.json"
)

// TodoStore Represents store of todo items
type TodoStore interface {
	// GetAllItems List all the items.
	// Returns a a collection of todo items
	GetAllItems() *internal.TodoCollection

	// GetItem Get a single todo item by it's unique id.
	// Returns a single todo item.
	GetItem(id int) *internal.Todo

	// AddItem Add a single item.
	// Returns a count of the amount of items added
	// this will be 1 or -1 indicating an error.
	AddItem(value string) int

	// DeleteItem Delete a single item by it's unique id.
	// Returns a count of the amount of items deleted
	// this will be 1 or -1 indicating an error.
	DeleteItem(ids ...int) int

	// DeleteAllItems Delete all items from the store.
	// Returns a count of the amount of items deleted.
	DeleteAllItems() int

	// EditItem Edit a single item.
	// Updated the item with the given id to the given value.
	// Returns the count of items updated
	EditItem(id int, value string) int
}

func Setup(useDb bool) (string, error) {

	homeDir := os.Getenv("HOME")

	if homeDir != "" {
		folder, e := setupFolder(homeDir)

		if e != nil {
			return "", e
		}

		if useDb {
			return folder + "/" + DB_FILE, nil
		}

		return folder + "/" + JSON_FILE, nil
	}

	return "", errors.New("Cannot find HOME directory.")
}

func setupFolder(homePath string) (string, error) {
	folder := fmt.Sprintf("%s/.todo", homePath)

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		e := os.Mkdir(folder, os.ModePerm)
		if e != nil {
			return "", e
		}
	}

	return folder, nil
}
