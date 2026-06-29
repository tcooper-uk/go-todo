package storage

import (
	"errors"
	"fmt"
	"github.com/tcooper-uk/go-todo/internal"
	"os"
)

type Mode uint8

const (
	FileMode  Mode = 0
	DbMode    Mode = 1
	CloudMode Mode = 2
)

const (
	DB_FILE   = "todo.db"
	JSON_FILE = "todo.json"
	KEY_FILE  = "firestore_key.json"
)

// ListOptions controls filtering for GetAllItems.
type ListOptions struct {
	ShowDone bool
	OnlyDone bool
	Priority string
	Tag      string
	Overdue  bool
}

// TodoStore Represents store of todo items
type TodoStore interface {
	// GetAllItems List all items, filtered by opts.
	GetAllItems(opts ListOptions) *internal.TodoCollection

	// GetItem Get a single todo item by its unique id.
	GetItem(id int) *internal.Todo

	// AddItem Add a single item. The store assigns ID and timestamps.
	// Returns 1 on success, 0 on error.
	AddItem(todo internal.Todo) int

	// DeleteItem Delete items by id.
	// Returns count deleted, or 0 on error.
	DeleteItem(ids ...int) int

	// DeleteAllItems Delete all items from the store.
	// Returns count deleted.
	DeleteAllItems() int

	// EditItem Update the item with the given id to match todo.
	// Caller should read the item first, mutate fields, then pass it back.
	// Returns count updated.
	EditItem(id int, todo internal.Todo) int
}

func Setup(mode Mode) (string, error) {

	homeDir := os.Getenv("HOME")

	if homeDir != "" {
		folder, e := setupFolder(homeDir)

		if e != nil {
			return "", e
		}

		switch mode {
		case DbMode:
			return folder + "/" + DB_FILE, nil
		case FileMode:
			return folder + "/" + JSON_FILE, nil
		case CloudMode:
			return findFirestoreKey(folder)
		}
	}

	return "", errors.New("Cannot find HOME directory.")
}

func findFirestoreKey(folder string) (string, error) {
	keyfile := folder + "/" + KEY_FILE
	if _, err := os.Stat(keyfile); errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	return keyfile, nil
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
