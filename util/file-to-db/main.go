package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tcooper-uk/go-todo/internal/storage"
)

func main() {

	dbPath, _ := storage.Setup(storage.DbMode)
	jsonFile, _ := storage.Setup(storage.FileMode)
	localStore := storage.NewLocalFileStore(jsonFile)
	collection := localStore.GetAllItems()

	db, _ := sql.Open("sqlite3", dbPath)
	tx, _ := db.Begin()

	stmt, _ := tx.Prepare("INSERT INTO todo_item (created_at, updated_at, name) VALUES (?, ?, ?)")
	defer stmt.Close()

	for _, item := range collection.Items {
		stmt.Exec(item.CreatedAt.UnixMilli(), item.UpdatedAt.UnixMilli(), item.Name)
	}

	tx.Commit()
}
