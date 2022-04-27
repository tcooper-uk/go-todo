package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tcooper-uk/go-todo/internal"
	s "github.com/tcooper-uk/go-todo/internal/storage"
)

func printItems(store s.TodoStore) {
	items := store.GetAllItems()

	const maxChars = 100

	// account for the ...
	if items.MaxLengthItem >= maxChars {
		items.MaxLengthItem = items.MaxLengthItem + 3
	}

	// print headers
	headerPadding := items.MaxLengthItem - 4
	if headerPadding < 0 {
		headerPadding = 0
	}

	fmt.Printf("ID\tItem%s\tCreated At\n", strings.Repeat(" ", headerPadding))

	// print items
	for _, item := range items.Items {

		i := utf8.RuneCountInString(item.Name)

		fmt.Printf("[%d]\t", item.ID)
		var printCount int
		for _, runeVal := range item.Name {
			if printCount == maxChars {
				fmt.Print("...")
				break
			}

			fmt.Printf("%c", runeVal)
			printCount++
		}

		remainder := items.MaxLengthItem - i
		fmt.Print(strings.Repeat(" ", remainder))

		fmt.Printf("\t%s\n", item.CreatedAt.Format("Mon 02 Jan 06 15:04"))
	}
}

func printItem(item *internal.Todo) {
	fmt.Printf("ID:\t\t%d\n", item.ID)
	fmt.Printf("Item:\t\t%s\n", item.Name)
	fmt.Printf("Created At:\t%s\n", item.CreatedAt.Format("Mon 02 Jan 06 15:04"))
	fmt.Printf("Updated At:\t%s\n", item.UpdatedAt.Format("Mon 02 Jan 06 15:04"))
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

	filePath := s.FILE_NAME

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
	argCount := len(args)

	store := s.NewLocalFileStore(determineFilePath())

	if argCount == 0 {
		printItems(store)
		return
	}

	// first argument is an ID
	if id, err := strconv.ParseInt(args[0], 0, 0); argCount == 1 && err == nil {
		item := store.GetItem(int(id))

		if item == nil {
			// cannot find item
			fmt.Printf("Cannot find item with ID %d\n", id)
		}
		printItem(item)
		return
	}

	switch args[0] {
	case "list", "l", "ps", "ls":
		printItems(store)
		return
	case "delete", "remove", "d", "rm":
		ids := parseIds(args[1:]...)
		store.DeleteItem(ids...)
	case "add", "create", "put", "a":
		value := strings.Join(args[1:], " ")
		store.AddItem(value)
	case "e", "edit", "update":

		if len(args) <= 1 {
			fmt.Println("You must supply and ID and new value.")
		}

		id := parseIds(args[1])
		value := strings.Join(args[2:], " ")
		count := store.EditItem(id[0], value)

		if count == 0 {
			fmt.Printf("Cannot find item with ID %d\n", id[0])
		}

	case "clearall":
		store.DeleteAllItems()
	default:
		fmt.Printf("Unknown command %s\n", args[0])
	}
}
