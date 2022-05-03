package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	t "github.com/tcooper-uk/go-todo/internal"
)

const FILE_NAME = "todo.json"

type LocalFileStore struct {
	items map[int]t.Todo

	FilePath  string
	ItemCount int
	MaxId     int
}

func NewLocalFileStore(filePath string) *LocalFileStore {

	items := make(map[int]t.Todo)

	// maps are refernece types, so no need for a pointer here
	// the method will fill the map
	size, maxId, _ := loadItemsFromFile(filePath, items)

	return &LocalFileStore{
		items:     items,
		FilePath:  filePath,
		ItemCount: size,
		MaxId:     maxId,
	}
}

func (store *LocalFileStore) GetAllItems() *t.TodoCollection {
	var results []t.Todo
	var count int
	var maxLength int
	for _, v := range store.items {
		count++

		valueLength := utf8.RuneCountInString(v.Name)
		if valueLength > maxLength {
			maxLength = valueLength
		}

		setUpdatedAtIfRequired(&v)

		results = append(results, v)
	}

	// sort by id
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return &t.TodoCollection{
		Items:         results,
		Size:          count,
		MaxLengthItem: maxLength,
	}

}

func (store *LocalFileStore) GetItem(id int) *t.Todo {
	item := store.items[id]
	setUpdatedAtIfRequired(&item)
	return &item
}

func (store *LocalFileStore) AddItem(value string) int {
	defer saveItems(store.FilePath, store.items)
	store.MaxId++
	store.items[store.MaxId] = *t.NewTodo(store.MaxId, value, time.Now())
	return 1
}

func (store *LocalFileStore) DeleteItem(ids ...int) int {
	defer saveItems(store.FilePath, store.items)

	var count int
	for _, id := range ids {
		count++
		if _, exists := store.items[id]; !exists {
			return -1
		}

		delete(store.items, id)
	}

	return count
}

func (store *LocalFileStore) DeleteAllItems() int {
	defer saveItems(store.FilePath, store.items)

	s := store.ItemCount
	store.items = make(map[int]t.Todo)
	store.ItemCount = 0
	return s
}

func (store *LocalFileStore) EditItem(id int, value string) int {
	defer saveItems(store.FilePath, store.items)
	i, exists := store.items[id]

	if !exists {
		return 0
	}

	i.Name = value
	i.UpdatedAt = time.Now()
	store.items[id] = i

	return 1
}

func setUpdatedAtIfRequired(item *t.Todo) {
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
}

func saveItems(path string, items map[int]t.Todo) {

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
	j := json.NewEncoder(file)
	if j.Encode(&tmp) != nil {
		fmt.Println("There was an error saving the todo list.", err)
	}
}

func loadItemsFromFile(path string, items map[int]t.Todo) (int, int, error) {

	loadErr := errors.New("Unable to load items from file.")

	f, err := getFile(path)
	defer f.Close()

	if err != nil {
		fmt.Println(loadErr, err)
		return 0, 0, loadErr
	}

	var tmp []t.Todo

	decode := json.NewDecoder(f)
	if decode.Decode(&tmp) != nil {
		return 0, 0, loadErr
	}

	var size, maxId int

	for _, item := range tmp {

		size++

		// add to map
		items[item.ID] = item

		// keep track of max id
		if item.ID > maxId {
			maxId = item.ID
		}
	}

	return size, maxId, nil
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
