package db

import (
	"cloud.google.com/go/firestore"
	"errors"
	"fmt"
	"github.com/tcooper-uk/go-todo/internal"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"sort"
	"time"
)

const (
	ProjectId  = "todo-de411"
	collection = "todos"
)

type CloudStore struct {
	client *firestore.Client
}

type CloudStoreConfig struct {
	ProjectId string
	KeyFile   string
	Client    *firestore.Client
}

func NewCloudStore(config *CloudStoreConfig) (*CloudStore, error) {

	if len(config.ProjectId) == 0 {
		return nil, errors.New("you must supply a firestore project id")
	}

	var client *firestore.Client
	var err error

	if config.Client == nil {
		ctx := context.Background()
		client, err = firestore.NewClient(ctx, config.ProjectId, option.WithCredentialsFile(config.KeyFile))
		if err != nil {
			return nil, fmt.Errorf("unable to create connection to cloudstore project %s, err: %w", config.ProjectId, err)
		}
	} else {
		client = config.Client
	}

	return &CloudStore{client: client}, err
}

// GetAllItems List all the items.
// Returns a collection of todo items
func (store *CloudStore) GetAllItems() *internal.TodoCollection {
	docRefs, err := store.client.CollectionGroup(collection).Query.
		Documents(context.Background()).
		GetAll()

	if err != nil {
		return &internal.TodoCollection{}
	}

	var items []internal.Todo
	var maxTitle = 0

	for _, each := range docRefs {
		var todo = internal.Todo{}
		err := each.DataTo(&todo)

		if err != nil {
			panic(err)
		}

		nameLen := len(todo.Name)
		if nameLen > maxTitle {
			maxTitle = nameLen
		}

		items = append(items, todo)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return &internal.TodoCollection{
		Items:         items,
		Size:          len(items),
		MaxLengthItem: maxTitle,
	}

}

// GetItem Get a single todo item by it's unique id.
// Returns a single todo item.
func (store *CloudStore) GetItem(id int) *internal.Todo {

	query := store.client.Collection(collection).Query.
		Where("ID", "==", id).
		Limit(1)

	// just assume one for now
	doc, err := query.Documents(context.Background()).Next()

	if err != nil {
		return nil
	}

	var todo = internal.Todo{}
	err = doc.DataTo(&todo)
	if err != nil {
		return nil
	}

	return &todo
}

// AddItem Add a single item.
// Returns a count of the amount of items added
// this will be 1 or -1 indicating an error.
func (store *CloudStore) AddItem(value string) int {

	nextId := 1
	now := time.Now()

	collectionRef := store.client.Collection(collection)
	documentIterator := collectionRef.Query.
		OrderBy("CreatedAt", firestore.Desc).
		Limit(1).
		Documents(context.Background())

	doc, err := documentIterator.Next()

	if err == nil {
		id := doc.Data()["ID"].(int64)
		nextId = int(id) + 1
	}

	td := internal.Todo{
		ID:        nextId,
		Name:      value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, _, err = collectionRef.Add(context.Background(), td)
	if err != nil {
		return 0
	}
	return 1
}

// DeleteItem Delete a single item by it's unique id.
// Returns a count of the amount of items deleted
// this will be 1 or -1 indicating an error.
func (store *CloudStore) DeleteItem(ids ...int) int {
	deleteCount := 0
	bulkWriter := store.client.BulkWriter(context.Background())
	query := store.client.Collection(collection).Query.
		Where("ID", "in", ids).
		Limit(1)

	iter := query.Documents(context.Background())

	for doc, err := iter.Next(); !errors.Is(err, iterator.Done); doc, err = iter.Next() {
		bulkWriter.Delete(doc.Ref)
		deleteCount++
	}

	bulkWriter.Flush()
	return deleteCount
}

// DeleteAllItems Delete all items from the store.
// Returns a count of the amount of items deleted.
func (store *CloudStore) DeleteAllItems() int {
	deleteCount := 0
	bulkWriter := store.client.BulkWriter(context.Background())
	doc, _ := store.client.Collection(collection).Documents(context.Background()).GetAll()

	for _, each := range doc {
		bulkWriter.Delete(each.Ref)
		deleteCount++
	}

	bulkWriter.Flush()
	return deleteCount
}

// EditItem Edit a single item.
// Updated the item with the given id to the given value.
// Returns the count of items updated
func (store *CloudStore) EditItem(id int, value string) int {
	query := store.client.Collection(collection).Query.
		Where("ID", "==", id).
		Limit(1)

	// just assume one for now
	doc, err := query.Documents(context.Background()).Next()

	if err != nil {
		return 0
	}

	update, err := doc.Ref.Update(context.Background(), []firestore.Update{
		{Path: "Name", Value: value},
	})

	if update != nil && err == nil {
		return 1
	} else {
		return 0
	}
}
