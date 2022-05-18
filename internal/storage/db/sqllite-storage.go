package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tcooper-uk/go-todo/internal"
)

const (
	setupStatement = `
		CREATE TABLE IF NOT EXISTS todo_item (
			id INTEGER PRIMARY KEY NOT NULL,
			created_at NUMERIC NOT NULL,
			updated_at NUMERIC NOT NULL,
			name TEXT NOT NULL
		);
	`
	fields = `id, created_at, updated_at, name`
)

type SQLLiteStore struct {
	db         *sql.DB
	DbFilePath string
}

type rowScan func(dest ...any) error

func NewSQLLiteStorage(dbPath string) (*SQLLiteStore, error) {

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_, err := os.Create(dbPath)
		if err != nil {
			return nil, dbErr(err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return nil, dbErr(err)
	}

	// migrate
	db.Exec(setupStatement)

	return &SQLLiteStore{
		db,
		dbPath,
	}, nil
}

func (store *SQLLiteStore) GetAllItems() *internal.TodoCollection {
	var items []internal.Todo

	rows, err := store.db.Query("SELECT " + fields + " FROM todo_item")

	if err != nil {
		return &internal.TodoCollection{}
	}

	defer rows.Close()
	count := mapToTodoItems(rows, &items)

	return &internal.TodoCollection{
		Items:         items,
		Size:          count,
		MaxLengthItem: findMaxLen(store.db),
	}
}

// Get a single todo item by it's unique id.
// Returns a single todo item.
func (store *SQLLiteStore) GetItem(id int) *internal.Todo {
	row := store.db.QueryRow("SELECT "+fields+" FROM todo_item WHERE id = ?", id)

	item, err := mapToTodoItem(row.Scan)

	if err != nil {
		return &internal.Todo{}
	}

	return item
}

// Add a single item.
// Returns a count of the amount of items added
// this will be 1 or -1 indicating an error.
func (store *SQLLiteStore) AddItem(value string) int {
	tx, _ := store.db.Begin()

	stmt, err := tx.Prepare(`
		INSERT INTO todo_item (created_at, updated_at, name)
		VALUES (?, ?, ?)
	`)
	defer stmt.Close()

	now := time.Now().UnixMilli()
	_, err = stmt.Exec(now, now, value)

	if err != nil {
		tx.Rollback()
		return 0
	}

	tx.Commit()
	return 1
}

// Delete a single item by it's unique id.
// Returns a count of the amount of items deleted
// this will be 1 or -1 indicating an error.
func (store *SQLLiteStore) DeleteItem(ids ...int) int {

	tx, err := store.db.Begin()

	stmt, err := tx.Prepare("DELETE FROM todo_item WHERE id = ?")
	defer stmt.Close()

	if err != nil {
		tx.Rollback()
		return 0
	}

	var successCount int

	for _, i := range ids {
		_, err := stmt.Exec(i)
		if err == nil {
			successCount++
		}
	}

	if successCount == 0 {
		tx.Rollback()
		return 0
	}

	tx.Commit()
	return successCount
}

// Delete all items from the store.
// Returns a count of the amount of items deleted.
func (store *SQLLiteStore) DeleteAllItems() int {
	tx, _ := store.db.Begin()
	res, err := tx.Exec("DELETE FROM todo_item")

	if err != nil {
		tx.Rollback()
		return 0
	}

	tx.Commit()

	i, err := res.RowsAffected()

	if err != nil {
		return 0
	}

	return int(i)
}

// Edit a single item.
// Updated the item with the given id to the given value.
// Returns the count of items updated
func (store *SQLLiteStore) EditItem(id int, value string) int {
	tx, _ := store.db.Begin()

	stmt, err := tx.Prepare(`
		UPDATE todo_item
		SET 
			name = ?,
			updated_at = ?
			WHERE id = ?
	`)
	defer stmt.Close()

	if err != nil {
		tx.Rollback()
		return 0
	}

	now := time.Now().UnixMilli()
	res, err := stmt.Exec(value, now, id)

	tx.Commit()
	i, err := res.RowsAffected()

	if err != nil {
		return 0
	}

	return int(i)
}

func findMaxLen(db *sql.DB) int {
	var len int
	row := db.QueryRow("SELECT LENGTH(name) name_len FROM todo_item ti ORDER BY name_len DESC LIMIT 1")
	row.Scan(&len)
	return len
}

func dbErr(err error) error {
	return fmt.Errorf("The path to the DB provided does not exists %e", err)
}

func mapToTodoItems(rows *sql.Rows, items *[]internal.Todo) int {
	var count int
	for rows.Next() {
		item, err := mapToTodoItem(rows.Scan)

		if err == nil {
			count++
			*items = append(*items, *item)
		}
	}
	return count
}

func mapToTodoItem(s rowScan) (*internal.Todo, error) {
	var id int
	var name string
	var createdAt int64
	var updatedAt int64

	err := s(&id, &createdAt, &updatedAt, &name)

	if err != nil {
		return nil, err
	}

	return &internal.Todo{
		ID:        id,
		CreatedAt: time.UnixMilli(createdAt),
		UpdatedAt: time.UnixMilli(updatedAt),
		Name:      name,
	}, nil
}
