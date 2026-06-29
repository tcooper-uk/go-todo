package storage_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tcooper-uk/go-todo/internal"
	"github.com/tcooper-uk/go-todo/internal/storage"
	"github.com/tcooper-uk/go-todo/internal/storage/db"
)

func TestCanGetItemsFromSqlLite(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)

	collection := store.GetAllItems(storage.ListOptions{ShowDone: true})

	assert.Equal(t, 2, collection.Size)
	assert.Equal(t, 11, collection.MaxLengthItem)
	assert.NotEmpty(t, collection.Items)

	if len(collection.Items) > 0 {
		assert.Equal(t, "Just a test", collection.Items[0].Name)
		assert.Equal(t, "more", collection.Items[1].Name)
	}
}

func TestDefaultListHidesDoneItems(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)

	// Mark item 1 as done.
	item := store.GetItem(1)
	item.Done = true
	store.EditItem(1, *item)

	collection := store.GetAllItems(storage.ListOptions{})
	assert.Equal(t, 1, collection.Size)
	assert.Equal(t, "more", collection.Items[0].Name)
}

func TestCanGetSingleItemFromDb(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)
	unixMilliTime := int64(1257894000000)

	item := store.GetItem(1)

	assert.NotNil(t, item)
	assert.Equal(t, "Just a test", item.Name)
	assert.Equal(t, unixMilliTime, item.CreatedAt.UnixMilli())
	assert.Equal(t, unixMilliTime, item.UpdatedAt.UnixMilli())
}

func TestCanAddItemFromDb(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)

	collection := store.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 2, collection.Size)

	added := store.AddItem(internal.Todo{Name: "New Item"})

	assert.Equal(t, 1, added)

	collection = store.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 3, collection.Size)
}

func TestCanAddItemWithFieldsFromDb(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)

	due := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	todo := internal.Todo{
		Name:     "tagged",
		Priority: internal.PriorityMedium,
		DueDate:  &due,
		Tags:     []string{"work", "later"},
	}
	store.AddItem(todo)

	collection := store.GetAllItems(storage.ListOptions{ShowDone: true})
	var got *internal.Todo
	for i := range collection.Items {
		if collection.Items[i].Name == "tagged" {
			got = &collection.Items[i]
		}
	}
	assert.NotNil(t, got)
	assert.Equal(t, internal.PriorityMedium, got.Priority)
	assert.Equal(t, []string{"work", "later"}, got.Tags)
	assert.NotNil(t, got.DueDate)
}

func TestCanDeleteItemFromDb(t *testing.T) {

	filePath, store := getStore(t)
	defer tearDown(filePath)

	collection := store.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 2, collection.Size)

	deleted := store.DeleteItem(1, 2)
	assert.Equal(t, 2, deleted)

	collection = store.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 0, collection.Size)
}

func TestCanDeleteAllItemsFromDb(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)

	collection := store.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 2, collection.Size)

	deleted := store.DeleteAllItems()
	assert.Equal(t, 2, deleted)

	collection = store.GetAllItems(storage.ListOptions{ShowDone: true})
	assert.Equal(t, 0, collection.Size)
}

func TestCanEditItemInDb(t *testing.T) {
	filePath, store := getStore(t)
	defer tearDown(filePath)

	item := store.GetItem(2)
	updatedAt := item.UpdatedAt
	assert.Equal(t, "more", item.Name)

	item.Name = "new"
	updated := store.EditItem(2, *item)
	assert.Equal(t, 1, updated)

	item = store.GetItem(2)
	assert.Equal(t, "new", item.Name)
	assert.Greater(t, item.UpdatedAt, updatedAt)
}

func getStore(t *testing.T) (string, *db.SQLLiteStore) {

	f, _ := os.CreateTemp(".", "*.db")
	defer f.Close()

	dbFile, _ := os.Open("test.db")
	defer dbFile.Close()

	io.Copy(f, dbFile)

	store, err := db.NewSQLLiteStorage(f.Name())
	assert.Nil(t, err)
	return f.Name(), store
}

func tearDown(filePath string) {
	os.Remove(filePath)
}
