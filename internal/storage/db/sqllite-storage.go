package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tcooper-uk/go-todo/internal"
	"github.com/tcooper-uk/go-todo/internal/storage"
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
	fields = `id, created_at, updated_at, name, done, priority, due_date, tags`
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

	db.Exec(setupStatement)

	if err := migrateSchema(db); err != nil {
		return nil, fmt.Errorf("schema migration failed: %w", err)
	}

	return &SQLLiteStore{
		db,
		dbPath,
	}, nil
}

func migrateSchema(db *sql.DB) error {
	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)

	if version >= 1 {
		return nil
	}

	migrations := []string{
		`ALTER TABLE todo_item ADD COLUMN done     INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE todo_item ADD COLUMN priority TEXT    NOT NULL DEFAULT ''`,
		`ALTER TABLE todo_item ADD COLUMN due_date NUMERIC`,
		`ALTER TABLE todo_item ADD COLUMN tags     TEXT    NOT NULL DEFAULT '[]'`,
		`PRAGMA user_version = 1`,
	}

	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (store *SQLLiteStore) GetAllItems(opts storage.ListOptions) *internal.TodoCollection {
	var items []internal.Todo

	query, args := buildListQuery(opts)
	rows, err := store.db.Query(query, args...)

	if err != nil {
		return &internal.TodoCollection{}
	}

	defer rows.Close()
	mapToTodoItems(rows, &items)

	// client-side tag filter
	if opts.Tag != "" {
		var filtered []internal.Todo
		for _, item := range items {
			if hasTag(item.Tags, opts.Tag) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	maxLen := 0
	for _, item := range items {
		if l := len(item.Name); l > maxLen {
			maxLen = l
		}
	}

	return &internal.TodoCollection{
		Items:         items,
		Size:          len(items),
		MaxLengthItem: maxLen,
	}
}

func buildListQuery(opts storage.ListOptions) (string, []any) {
	base := "SELECT " + fields + " FROM todo_item"
	var conditions []string
	var args []any

	if !opts.ShowDone && !opts.OnlyDone {
		conditions = append(conditions, "done = 0")
	} else if opts.OnlyDone {
		conditions = append(conditions, "done = 1")
	}

	if opts.Priority != "" {
		conditions = append(conditions, "priority = ?")
		args = append(args, opts.Priority)
	}

	if opts.Overdue {
		conditions = append(conditions, "due_date IS NOT NULL AND due_date < ?")
		args = append(args, time.Now().UnixMilli())
	}

	if len(conditions) > 0 {
		base += " WHERE " + strings.Join(conditions, " AND ")
	}

	return base, args
}

// GetItem Get a single todo item by its unique id.
func (store *SQLLiteStore) GetItem(id int) *internal.Todo {
	row := store.db.QueryRow("SELECT "+fields+" FROM todo_item WHERE id = ?", id)

	item, err := mapToTodoItem(row.Scan)

	if err != nil {
		return &internal.Todo{}
	}

	return item
}

// AddItem Add a single item. The store assigns ID and timestamps.
func (store *SQLLiteStore) AddItem(todo internal.Todo) int {
	tx, err := store.db.Begin()
	if err != nil {
		return 0
	}

	stmt, err := tx.Prepare(`
		INSERT INTO todo_item (created_at, updated_at, name, done, priority, due_date, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return 0
	}
	defer stmt.Close()

	now := time.Now().UnixMilli()
	tagsJSON, _ := json.Marshal(todo.Tags)
	var dueDateVal any
	if todo.DueDate != nil {
		dueDateVal = todo.DueDate.UnixMilli()
	}

	_, err = stmt.Exec(now, now, todo.Name, boolToInt(todo.Done), string(todo.Priority), dueDateVal, string(tagsJSON))

	if err != nil {
		tx.Rollback()
		return 0
	}

	tx.Commit()
	return 1
}

// DeleteItem Delete items by id.
func (store *SQLLiteStore) DeleteItem(ids ...int) int {

	tx, err := store.db.Begin()
	if err != nil {
		return 0
	}

	stmt, err := tx.Prepare("DELETE FROM todo_item WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return 0
	}
	defer stmt.Close()

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

// DeleteAllItems Delete all items from the store.
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

// EditItem Update the item with the given id.
func (store *SQLLiteStore) EditItem(id int, todo internal.Todo) int {
	tx, err := store.db.Begin()
	if err != nil {
		return 0
	}

	stmt, err := tx.Prepare(`
		UPDATE todo_item
		SET name = ?, updated_at = ?, done = ?, priority = ?, due_date = ?, tags = ?
		WHERE id = ?
	`)
	if err != nil {
		tx.Rollback()
		return 0
	}
	defer stmt.Close()

	now := time.Now().UnixMilli()
	tagsJSON, _ := json.Marshal(todo.Tags)
	var dueDateVal any
	if todo.DueDate != nil {
		dueDateVal = todo.DueDate.UnixMilli()
	}

	res, err := stmt.Exec(todo.Name, now, boolToInt(todo.Done), string(todo.Priority), dueDateVal, string(tagsJSON), id)
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

func findMaxLen(db *sql.DB) int {
	var l int
	row := db.QueryRow("SELECT LENGTH(name) name_len FROM todo_item ti ORDER BY name_len DESC LIMIT 1")
	row.Scan(&l)
	return l
}

func dbErr(err error) error {
	return fmt.Errorf("The path to the DB provided does not exists %e", err)
}

func mapToTodoItems(rows *sql.Rows, items *[]internal.Todo) {
	for rows.Next() {
		item, err := mapToTodoItem(rows.Scan)

		if err == nil {
			*items = append(*items, *item)
		}
	}
}

func mapToTodoItem(s rowScan) (*internal.Todo, error) {
	var id int
	var name string
	var createdAt int64
	var updatedAt int64
	var done int
	var priority string
	var dueDateMs sql.NullInt64
	var tagsJSON string

	err := s(&id, &createdAt, &updatedAt, &name, &done, &priority, &dueDateMs, &tagsJSON)

	if err != nil {
		return nil, err
	}

	var tags []string
	if tagsJSON != "" && tagsJSON != "null" {
		json.Unmarshal([]byte(tagsJSON), &tags)
	}

	var dueDate *time.Time
	if dueDateMs.Valid {
		t := time.UnixMilli(dueDateMs.Int64)
		dueDate = &t
	}

	return &internal.Todo{
		ID:        id,
		CreatedAt: time.UnixMilli(createdAt),
		UpdatedAt: time.UnixMilli(updatedAt),
		Name:      name,
		Done:      done != 0,
		Priority:  internal.Priority(priority),
		DueDate:   dueDate,
		Tags:      tags,
	}, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func hasTag(tags []string, tag string) bool {
	for _, tg := range tags {
		if tg == tag {
			return true
		}
	}
	return false
}
