package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	s "github.com/tcooper-uk/go-todo/internal/storage"
)

var store s.TodoStore

func printItems() {
	items := store.GetAllItems()

	const maxChars = 100
	var maxLen int

	for _, item := range items {
		i := utf8.RuneCountInString(item.Name)

		if i >= maxChars {
			maxLen = maxChars + 3
			break
		}

		if i > maxLen {
			maxLen = i
		}
	}

	for _, item := range items {

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

		remainder := maxLen - i
		for i := 0; i < remainder; i++ {
			fmt.Print(" ")
		}

		fmt.Printf("\t%s\n", item.CreatedAt.Format("Mon 02 Jan 06 15:04"))

		// fmt.Printf("[%d]\t%s\t%s\n",
		// 	item.ID,
		// 	item.Name,
		// 	item.CreatedAt.Format("Mon 02 Jan 06 15:04"),
		// )
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
	case "e", "edit":

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
