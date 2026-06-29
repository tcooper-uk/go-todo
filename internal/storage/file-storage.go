package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	t "github.com/tcooper-uk/go-todo/internal"
)

type LocalFileStore struct {
	items map[int]t.Todo

	FilePath  string
	ItemCount int
	MaxId     int
}

func NewLocalFileStore(filePath string) *LocalFileStore {

	items := make(map[int]t.Todo)

	size, maxId, _ := loadItemsFromFile(filePath, items)

	return &LocalFileStore{
		items:     items,
		FilePath:  filePath,
		ItemCount: size,
		MaxId:     maxId,
	}
}

func (store *LocalFileStore) GetAllItems(opts ListOptions) *t.TodoCollection {
	var results []t.Todo
	var maxLength int
	now := time.Now()

	for _, v := range store.items {
		setUpdatedAtIfRequired(&v)

		if !opts.ShowDone && !opts.OnlyDone && v.Done {
			continue
		}
		if opts.OnlyDone && !v.Done {
			continue
		}
		if opts.Priority != "" && string(v.Priority) != opts.Priority {
			continue
		}
		if opts.Overdue && (v.DueDate == nil || !v.DueDate.Before(now)) {
			continue
		}
		if opts.Tag != "" && !hasTag(v.Tags, opts.Tag) {
			continue
		}

		nameLen := utf8.RuneCountInString(v.Name)
		if nameLen > maxLength {
			maxLength = nameLen
		}

		results = append(results, v)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	return &t.TodoCollection{
		Items:         results,
		Size:          len(results),
		MaxLengthItem: maxLength,
	}
}

func (store *LocalFileStore) GetItem(id int) *t.Todo {
	item, exists := store.items[id]
	if !exists {
		return nil
	}
	setUpdatedAtIfRequired(&item)
	return &item
}

func (store *LocalFileStore) AddItem(todo t.Todo) int {
	defer saveItems(store.FilePath, store.items)
	store.MaxId++
	now := time.Now()
	todo.ID = store.MaxId
	todo.CreatedAt = now
	todo.UpdatedAt = now
	store.items[store.MaxId] = todo
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

	s := len(store.items)
	store.items = make(map[int]t.Todo)
	store.ItemCount = 0
	return s
}

func (store *LocalFileStore) EditItem(id int, todo t.Todo) int {
	defer saveItems(store.FilePath, store.items)
	if _, exists := store.items[id]; !exists {
		return 0
	}

	todo.ID = id
	todo.UpdatedAt = time.Now()
	store.items[id] = todo

	return 1
}

func hasTag(tags []string, tag string) bool {
	for _, tg := range tags {
		if tg == tag {
			return true
		}
	}
	return false
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

	file.Truncate(0)
	file.Seek(0, 0)

	var tmp []t.Todo
	for _, v := range items {
		tmp = append(tmp, v)
	}

	j := json.NewEncoder(file)
	if j.Encode(&tmp) != nil {
		fmt.Println("There was an error saving the todo list.", err)
	}
}

func loadItemsFromFile(path string, items map[int]t.Todo) (int, int, error) {

	loadErr := errors.New("unable to load items from file")

	f, err := getFile(path)
	if err != nil {
		fmt.Println(loadErr, err)
		return 0, 0, loadErr
	}
	defer f.Close()

	var tmp []t.Todo

	decode := json.NewDecoder(f)
	if decode.Decode(&tmp) != nil {
		return 0, 0, loadErr
	}

	var size, maxId int

	for _, item := range tmp {

		size++

		items[item.ID] = item

		if item.ID > maxId {
			maxId = item.ID
		}
	}

	return size, maxId, nil
}

func getFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}
	return f, nil
}
