package main

import (
	"github.com/tcooper-uk/go-todo/internal/storage"
	"github.com/tcooper-uk/go-todo/internal/storage/db"
)

func main() {
	dbPath, _ := storage.Setup(storage.DbMode)
	firestorePath, _ := storage.Setup(storage.CloudMode)
	firestore, _ := db.NewCloudStore(&db.CloudStoreConfig{
		ProjectId: "todo-de411",
		KeyFile:   firestorePath,
	})

	liteStorage, _ := db.NewSQLLiteStorage(dbPath)
	items := liteStorage.GetAllItems()

	firestore.DeleteAllItems()

	for _, item := range items.Items {
		firestore.AddItem(item.Name)
	}
}
