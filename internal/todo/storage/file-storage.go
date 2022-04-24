package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	t "github.com/tcooper-uk/go-todo/internal/todo"
)

const FILE_NAME = "todo.json"

// in mempry store of items
var items map[int]t.Todo = make(map[int]t.Todo)
var size, maxId int

type LocalFileStore struct {
	FilePath  string
	ItemCount int
	MaxId     int
}

func NewLocalFileStore(filePath string) *LocalFileStore {
	loadItemsFromFile(filePath)
	return &LocalFileStore{
		FilePath:  filePath,
		ItemCount: size,
		MaxId:     maxId,
	}
}

func (store *LocalFileStore) GetAllItems() []t.Todo {
	var results []t.Todo

	for _, v := range items {
		results = append(results, v)
	}

	return results
}

func (store *LocalFileStore) GetItem(id int) t.Todo {
	return items[id]
}

func (store *LocalFileStore) AddItem(value string) int {
	defer saveItems(store.FilePath)
	maxId++
	items[maxId] = *t.NewTodo(maxId, value, time.Now())
	return 1
}

func (store *LocalFileStore) DeleteItem(ids ...int) int {
	defer saveItems(store.FilePath)

	var count int
	for _, id := range ids {
		count++
		if _, exists := items[id]; !exists {
			return -1
		}

		delete(items, id)
	}

	return count
}

func (store *LocalFileStore) DeleteAllItems() int {
	defer saveItems(store.FilePath)

	s := size
	items = make(map[int]t.Todo)
	size = 0
	return s
}

func saveItems(path string) {

	file, err := getFile(path)

	if err != nil {
		fmt.Println(err)
	}

	// clear the file of the old contents
	file.Truncate(0)
	file.Seek(0, 0)

	var tmp []t.Todo
	for _, v := range items {
		tmp = append(tmp, v)
	}

	// build json and write out
	json, err := json.Marshal(tmp)
	_, err = file.Write(json)

	if err != nil {
		fmt.Println("There was an error saving the todo list.", err)
	}
}

func loadItemsFromFile(path string) error {

	loadErr := errors.New("Unable to load items from file.")

	f, err := getFile(path)
	defer f.Close()

	if err != nil {
		fmt.Println(loadErr, err)
		return loadErr
	}

	var tmp []t.Todo

	stat, _ := f.Stat()
	buffer := make([]byte, stat.Size())
	n, _ := f.Read(buffer)

	if n == 0 || json.Unmarshal(buffer, &tmp) != nil {
		return loadErr
	}

	for _, item := range tmp {

		size++

		// add to map
		items[item.ID] = item

		// keep track of max id
		if item.ID > maxId {
			maxId = item.ID
		}
	}

	return nil
}

// get the storage file
func getFile(path string) (*os.File, error) {

	var filePath string

	if path != "" {
		filePath = path
	} else {
		ex, err := os.Executable()

		// use json file stored with binary
		if err == nil {
			exePath := filepath.Dir(ex)
			if !strings.HasSuffix(exePath, "/") {
				exePath = exePath + "/"
			}

			filePath = exePath + FILE_NAME
		}
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)

	if err != nil {
		return nil, err
	}

	return f, nil
}
