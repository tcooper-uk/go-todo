package storage_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcooper-uk/go-todo/internal/storage"
)

func TestCanGetAllItems(t *testing.T) {
	s := storage.NewLocalFileStore("testtodo.json")

	items := s.GetAllItems()

	assert.Equal(t, 2, items.Size)
	assert.Equal(t, items.Items[0].ID, 1)
	assert.Equal(t, items.Items[0].Name, "first")
	assert.Equal(t, items.Items[1].ID, 2)
	assert.Equal(t, items.Items[1].Name, "second")
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

	initalItems := s.GetAllItems()

	s.AddItem("new item")

	items := s.GetAllItems()

	assert.Empty(t, initalItems)
	assert.NotEmpty(t, items)
	assert.Equal(t, items.Items[0].ID, 1)
	assert.Equal(t, items.Items[0].Name, "new item")
}

func TestCanEditItem(t *testing.T) {
	const filename = "newfile.json"
	defer cleanUp(filename)

	s := storage.NewLocalFileStore(filename)

	s.AddItem("value1")
	s.EditItem(1, "value2")

	items := s.GetAllItems()

	assert.NotEmpty(t, items)
	assert.Equal(t, items.Items[0].ID, 1)
	assert.Equal(t, items.Items[0].Name, "value2")
}

func TestCanDeleteItem(t *testing.T) {
	const filename = "deleteFile.json"
	defer cleanUp(filename)

	os.WriteFile(filename, []byte("[{\"id\": 1,\"name\": \"first\",\"created_at\": \"2022-04-20T14:32:28.901094+01:00\"}]"), 0755)

	s := storage.NewLocalFileStore(filename)

	items := s.GetAllItems()
	assert.NotEmpty(t, items)

	s.DeleteItem(1)

	items = s.GetAllItems()
	assert.Empty(t, items)
}

func TestCanDeleteAllItems(t *testing.T) {
	const filename = "deleteAllFile.json"
	defer cleanUp(filename)

	os.WriteFile(filename, []byte("[{\"id\": 1,\"name\": \"first\",\"created_at\": \"2022-04-20T14:32:28.901094+01:00\"}]"), 0755)

	s := storage.NewLocalFileStore(filename)

	items := s.GetAllItems()
	assert.NotEmpty(t, items)

	s.DeleteAllItems()

	items = s.GetAllItems()
	assert.Empty(t, items)
}

func cleanUp(file string) {
	os.Remove(file)
}
