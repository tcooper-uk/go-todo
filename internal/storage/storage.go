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
