package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type todo struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

var counter int
var items []todo

func NewTodo(counter int, name string, createdAt time.Time) todo {
	return todo{counter, name, createdAt}
}

func saveItems(file *os.File) {

	// clear the file of the old contents
	file.Truncate(0)
	file.Seek(0, 0)

	// build json and write out
	var err error
	json, err := json.Marshal(items)
	_, err = file.Write(json)

	if err != nil {
		log.Println("There was an error saving the todo list.")
	}
}

func readTodoItems(f *os.File) int {
	stat, _ := f.Stat()
	buffer := make([]byte, stat.Size())
	n, _ := f.Read(buffer)

	if n == 0 || json.Unmarshal(buffer, &items) != nil {
		return 0
	}

	topId := 0
	for _, item := range items {
		if item.ID > topId {
			topId = item.ID
		}
	}

	return topId
}

func getFile() *os.File {

	exPath := "todo.json"
	ex, err := os.Executable()

	// use json file stored with binary
	if err == nil {
		exPath = filepath.Dir(ex) + "/todo.json"
	}

	f, err := os.OpenFile(exPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)

	if err != nil {
		log.Println(err)
	}

	return f
}

func printItems() {
	for _, item := range items {
		fmt.Printf("[%d]\t%s\t%s\n",
			item.ID,
			item.Name,
			item.CreatedAt.Format("Mon Jan _2 15:04:05"),
		)
	}
}

func deleteItem(ids ...int) {
	for _, id := range ids {
		for index, item := range items {
			if item.ID == id {
				items = append(items[:index], items[index+1:]...)
			}
		}
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

func addItem(args ...string) {
	name := strings.Join(args, " ")
	counter++
	t := NewTodo(counter, name, time.Now())
	items = append(items, t)
}

func main() {

	// args without program name
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Not sure what you want.")
		return
	}

	file := getFile()
	defer file.Close()

	counter = readTodoItems(file)

	switch args[0] {
	case "list", "l", "ps", "ls":
		printItems()
	case "delete", "remove", "d":
		ids := parseIds(args[1:]...)
		deleteItem(ids...)
	case "add", "create", "put", "a":
		addItem(args[1:]...)
	default:
		fmt.Printf("Unknown command %s\n", args[0])
	}

	saveItems(file)
}
