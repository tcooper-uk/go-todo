package storage_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tcooper-uk/go-todo/internal"
	"github.com/tcooper-uk/go-todo/internal/storage"
)

func newTodo(name string) internal.Todo {
	return internal.Todo{Name: name}
}

func TestCanGetAllItems(t *testing.T) {
	s := storage.NewLocalFileStore("testtodo.json")

	items := s.GetAllItems(storage.ListOptions{ShowDone: true})

	assert.Equal(t, 2, items.Size)
	assert.Equal(t, items.Items[0].ID, 1)
	assert.Equal(t, items.Items[0].Name, "first")
	assert.Equal(t, items.Items[1].ID, 2)
	assert.Equal(t, items.Items[1].Name, "second")
}

func TestGetAllItemsHidesDoneByDefault(t *testing.T) {
	const filename = "filterdone.json"
	defer cleanUp(filename)

	s := storage.NewLocalFileStore(filename)
	s.AddItem(newTodo("active"))
	doneTodo := internal.Todo{Name: "done item", Done: true}
	s.AddItem(doneTodo)

	items := s.GetAllItems(storage.ListOptions{})
	assert.Equal(t, 1, items.Size)
	assert.Equal(t, "active", items.Items[0].Name)
}

func TestGetAllItemsOnlyDone(t *testing.T) {
	const filename = "onlydone.json"
	defer cleanUp(filename)

	s := storage.NewLocalFileStore(filename)
	s.AddItem(newTodo("active"))
	s.AddItem(internal.Todo{Name: "done item", Done: true})

	items := s.GetAllItems(storage.ListOptions{OnlyDone: true})
	assert.Equal(t, 1, items.Size)
	assert.Equal(t, "done item", items.Items[0].Name)
}

func TestCanGetSingleItem(t *testing.T) {
	s := storage.NewLocalFileStore("testtodo.json")

	item := s.GetItem(2)

	assert.Equal(t, item.ID, 2)
	assert.Equal(t, item.Name, "second")
}

func TestCanCreateNewItem(t *testing.T) {
	const filename = "newfile.json"
	defer cleanUp(filename)

	s := storage.NewLocalFileStore(filename)

	initalItems := s.GetAllItems(storage.ListOptions{ShowDone: true})

	s.AddItem(newTodo("new item"))

	items := s.GetAllItems(storage.ListOptions{ShowDone: true})

	assert.Empty(t, initalItems)
	assert.NotEmpty(t, items)
	assert.Equal(t, items.Items[0].ID, 1)
	assert.Equal(t, items.Items[0].Name, "new item")
}

func TestCanCreateItemWithFields(t *testing.T) {
	const filename = "newfields.json"
	defer cleanUp(filename)

	due := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	todo := internal.Todo{
		Name:     "rich item",
		Priority: internal.PriorityHigh,
		DueDate:  &due,
		Tags:     []string{"work", "urgent"},
	}

	s := storage.NewLocalFileStore(filename)
	s.AddItem(todo)

	items := s.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 1, items.Size)
	got := items.Items[0]
	assert.Equal(t, "rich item", got.Name)
	assert.Equal(t, internal.PriorityHigh, got.Priority)
	assert.Equal(t, []string{"work", "urgent"}, got.Tags)
	assert.NotNil(t, got.DueDate)
	assert.Equal(t, due.UTC(), got.DueDate.UTC())
}

func TestCanEditItem(t *testing.T) {
	const filename = "newfile.json"
	defer cleanUp(filename)

	s := storage.NewLocalFileStore(filename)

	s.AddItem(newTodo("value1"))

	item := s.GetItem(1)
	item.Name = "value2"
	s.EditItem(1, *item)

	items := s.GetAllItems(storage.ListOptions{ShowDone: true})

	assert.NotEmpty(t, items)
	assert.Equal(t, items.Items[0].ID, 1)
	assert.Equal(t, items.Items[0].Name, "value2")
}

func TestCanDeleteItem(t *testing.T) {
	const filename = "deleteFile.json"
	defer cleanUp(filename)

	os.WriteFile(filename, []byte("[{\"id\": 1,\"name\": \"first\",\"created_at\": \"2022-04-20T14:32:28.901094+01:00\"}]"), 0755)

	s := storage.NewLocalFileStore(filename)

	items := s.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.NotEmpty(t, items)

	s.DeleteItem(1)

	items = s.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Empty(t, items)
}

func TestCanDeleteAllItems(t *testing.T) {
	const filename = "deleteAllFile.json"
	defer cleanUp(filename)

	os.WriteFile(filename, []byte("[{\"id\": 1,\"name\": \"first\",\"created_at\": \"2022-04-20T14:32:28.901094+01:00\"}]"), 0755)

	s := storage.NewLocalFileStore(filename)

	items := s.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.NotEmpty(t, items)

	s.DeleteAllItems()

	items = s.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Empty(t, items)
}

func cleanUp(file string) {
	os.Remove(file)
}
