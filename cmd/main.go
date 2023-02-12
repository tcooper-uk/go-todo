package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tcooper-uk/go-todo/internal"
	"github.com/tcooper-uk/go-todo/internal/storage"
	s "github.com/tcooper-uk/go-todo/internal/storage"
	"github.com/tcooper-uk/go-todo/internal/storage/db"
)

const useDb = true

func main() {

	// args without program name
	args := os.Args[1:]
	argCount := len(args)

	filePath, err := storage.Setup(useDb)

	if err != nil {
		fmt.Println("Error:", err)
	}

	var store s.TodoStore
	if useDb {
		store, err = db.NewSQLLiteStorage(filePath)
	} else {
		store = s.NewLocalFileStore(filePath)
	}

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

		idErr := func() {
			fmt.Println("You must supply and ID and new value.")
			os.Exit(1)
		}

		if len(args) <= 1 {
			idErr()
		}

		id := parseIds(args[1])
		if len(id) == 0 {
			idErr()
		}

		value := strings.Join(args[2:], " ")
		count := store.EditItem(id[0], value)

		if count == 0 {
			fmt.Printf("Cannot find item with ID %d\n", id[0])
		}

	case "clearall":
		store.DeleteAllItems()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command %s\n", args[0])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Todo Store")
	fmt.Println("USAGE: todo [COMMAND] [ARGUMENT]")
	fmt.Println()
	fmt.Println("Commands")
	fmt.Printf("\tlist, l, ps, ls \t- list your todo items\n")

	fmt.Printf("\tdelete, remove, d, rm \t- delete a todo item but id.\n")
	fmt.Printf("\t\t- this will take the id as an arugmnet\n")

	fmt.Printf("\tadd, create, put, a \t- add a new item to the todo list.\n")
	fmt.Printf("\t\t- This will take the full title of the todo item as the arguments following the command.\n")

	fmt.Printf("\te, edit, update \t- edit an existing todo by the id.\n")
	fmt.Printf("\t\t- this will take the id as an arugmnet\n")

	fmt.Printf("\tclearall \t- delete all todo items.\n")

	fmt.Printf("\thelp, -h, --help \t- show this help text.\n")
}

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
