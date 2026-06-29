package db

import (
	"cloud.google.com/go/firestore"
	"errors"
	"fmt"
	"github.com/tcooper-uk/go-todo/internal"
	"github.com/tcooper-uk/go-todo/internal/storage"
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

// GetAllItems List all the items filtered by opts.
func (store *CloudStore) GetAllItems(opts storage.ListOptions) *internal.TodoCollection {
	docRefs, err := store.client.CollectionGroup(collection).Query.
		Documents(context.Background()).
		GetAll()

	if err != nil {
		return &internal.TodoCollection{}
	}

	now := time.Now()
	var items []internal.Todo
	var maxTitle = 0

	for _, each := range docRefs {
		var todo = internal.Todo{}
		err := each.DataTo(&todo)
		if err != nil {
			continue
		}

		if !opts.ShowDone && !opts.OnlyDone && todo.Done {
			continue
		}
		if opts.OnlyDone && !todo.Done {
			continue
		}
		if opts.Priority != "" && string(todo.Priority) != opts.Priority {
			continue
		}
		if opts.Overdue && (todo.DueDate == nil || !todo.DueDate.Before(now)) {
			continue
		}
		if opts.Tag != "" && !hasTag(todo.Tags, opts.Tag) {
			continue
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

// GetItem Get a single todo item by its unique id.
func (store *CloudStore) GetItem(id int) *internal.Todo {

	query := store.client.Collection(collection).Query.
		Where("ID", "==", id).
		Limit(1)

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

// AddItem Add a single item. The store assigns ID and timestamps.
func (store *CloudStore) AddItem(todo internal.Todo) int {

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

	todo.ID = nextId
	todo.CreatedAt = now
	todo.UpdatedAt = now

	_, _, err = collectionRef.Add(context.Background(), todo)
	if err != nil {
		return 0
	}
	return 1
}

// DeleteItem Delete items by id.
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

// EditItem Update the item with the given id to match todo.
func (store *CloudStore) EditItem(id int, todo internal.Todo) int {
	query := store.client.Collection(collection).Query.
		Where("ID", "==", id).
		Limit(1)

	doc, err := query.Documents(context.Background()).Next()

	if err != nil {
		return 0
	}

	now := time.Now()
	updates := []firestore.Update{
		{Path: "Name", Value: todo.Name},
		{Path: "Done", Value: todo.Done},
		{Path: "Priority", Value: todo.Priority},
		{Path: "DueDate", Value: todo.DueDate},
		{Path: "Tags", Value: todo.Tags},
		{Path: "UpdatedAt", Value: now},
	}

	result, err := doc.Ref.Update(context.Background(), updates)

	if result != nil && err == nil {
		return 1
	}
	return 0
}
