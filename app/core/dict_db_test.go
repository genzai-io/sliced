package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/genzai-io/sliced/common/btrdb"
	"github.com/genzai-io/sliced/proto/store"
)

func createStoreDB(t *testing.T) *databaseStore {
	// Create a memory DB
	db, err := btrdb.Open("store.db")
	if err != nil {
		t.Fatal(err)
	}

	d := newDatabases(db)
	if err := d.Start(); err != nil {
		t.Fatal(d)
	}

	return d
}

func TestStore_SelectDatabases(t *testing.T) {
	db := createStoreDB(t)
	databases, err := db.selectAll(db.tblDatabases.Pk)
	if err != nil {
		t.Fatal(err)
	}
	for _, database := range databases {
		data, err := json.Marshal(database)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(data))
	}
}

func TestStore_SelectDatabasesByName(t *testing.T) {
	db := createStoreDB(t)
	databases, err := db.selectAll(db.ByNames)
	if err != nil {
		t.Fatal(err)
	}
	for _, database := range databases {
		data, err := json.Marshal(database)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(data))
	}
}

func TestStore_UpdateDatabases(t *testing.T) {
	db := createStoreDB(t)

	err := db.db.Update(func(tx *btrdb.Tx) error {
		// Get it from the database
		value, err := db.tblDatabases.Get(tx, 2)
		if err != nil {
			return err
		}

		// Cast it
		database, ok := value.(*store.Database)
		if !ok {
			return btrdb.ErrUnexpectedDocument
		}

		// Modify it
		database.Name = "orders"

		// Update it in the database
		_, _, err = db.tblDatabases.Update(tx, database)
		return err
	})

	if err != nil {
		t.Fatal(err)
	}

	//databases, err := db.Select(db, TblDatabase.Pk)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//for _, database := range databases {
	//	data, err := json.Marshal(database)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	fmt.Println(string(data))
	//}
}

func TestStore_CreateDatabase(t *testing.T) {
	db := createStoreDB(t)

	num := 20
	created := make([]*store.Database, 0)
	for i := 0; i < num; i++ {
		database, err := db.Insert(fmt.Sprintf("app-%d", i+1))
		if err != nil {
			t.Fatal(err)
		}
		created = append(created, database)
	}

	//databases, err := Select(db)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if len(created) != len(databases) {
	//	t.Fatal(fmt.Errorf("created size: %d does not match selected size: %d", len(created), len(databases)))
	//}
	//for i, selected := range databases {
	//	cdb := created[i]
	//
	//	data1, err := selected.Marshal()
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	data2, err := cdb.Marshal()
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	if bytes.Compare(data1, data2) != 0 {
	//		t.Fatal(fmt.Errorf("%s != %s", selected, cdb))
	//	}
	//
	//	//fmt.Println(selected)
	//}
}

func TestStore_CreateDatabaseDuplicateName(t *testing.T) {
	db := createStoreDB(t)

	database, err := db.Insert("app")
	if err != nil {
		t.Fatal(err)
	}
	_ = database
	//fmt.Println(database)

	database2, err := db.Insert("app")
	if err == nil {
		t.Fatal(errors.New("no error... expected ErrExists"))
	}
	if err != nil && err != btrdb.ErrDuplicateKey {
		t.Fatal(err)
	}
	if database2 != nil {
		t.Fatal("expected database2 to be nil")
	}
}

func TestStore_CreateDatabaseIDSequence(t *testing.T) {
	db := createStoreDB(t)

	database, err := db.Insert("app")
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(database)

	database2, err := db.Insert("app-2")
	if err != nil {
		t.Fatal(err)
	}
	if database2.Id != database.Id+1 {
		t.Fatal(fmt.Errorf("expected second database to be assigned id: %d - instead it was: %d", database.Id+1, database2.Id))
	}
	//fmt.Println(database2)
}
