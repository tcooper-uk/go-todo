package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	s "github.com/tcooper-uk/go-todo/internal/todo/storage"
)

var store s.TodoStore

func printItems() {
	items := store.GetAllItems()
	for _, item := range items {
		fmt.Printf("[%d]\t%s\t%s\n",
			item.ID,
			item.Name,
			item.CreatedAt.Format("Mon 02 Jan 06 15:04"),
		)
	}
}

func parseIds(possibleIds ...string) []int {
	var ids []int

	for _, s := range possibleIds {
		id, err := strconv.Atoi(s)
		if err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

func determineFilePath() string {
	var filePath string

	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		checkPath := homeDir + "/" + s.FILE_NAME
		if _, err := os.Stat(checkPath); errors.Is(err, os.ErrNotExist) {
			return filePath
		} else {
			return checkPath
		}
	}
	return filePath
}

func main() {

	// args without program name
	args := os.Args[1:]

	store = s.NewLocalFileStore(determineFilePath())

	if len(args) == 0 {
		printItems()
		return
	}

	switch args[0] {
	case "list", "l", "ps", "ls":
		printItems()
		return
	case "delete", "remove", "d":
		ids := parseIds(args[1:]...)
		store.DeleteItem(ids...)
	case "add", "create", "put", "a":
		value := strings.Join(args[1:], " ")
		store.AddItem(value)
	case "clearall":
		store.DeleteAllItems()
	default:
		fmt.Printf("Unknown command %s\n", args[0])
	}
}
