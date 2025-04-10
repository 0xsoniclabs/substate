package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestBaseDB_FactorySuccess(t *testing.T) {
	dbPath := t.TempDir() + "test-db"

	db, err := NewDefaultBaseDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.NoError(t, db.Close())

	db, err = NewBaseDB(dbPath, nil, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.NoError(t, db.Close())

	db, err = NewReadOnlyBaseDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	err = db.Put([]byte{}, []byte{})
	assert.Error(t, leveldb.ErrReadOnly, err)
	assert.NoError(t, db.Close())

	db, err = OpenBaseDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.NoError(t, db.Close())
}

func TestBaseDB_FactoryError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultBaseDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err = NewDefaultBaseDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewBaseDB(dbPath, nil, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewReadOnlyBaseDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = OpenBaseDB("@@@_$$$__@@@")
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestBaseDB_MakeDefaultBaseDBFromBaseDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultBaseDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db2 := MakeDefaultBaseDBFromBaseDB(db)
	assert.NotNil(t, db2)
	assert.Equal(t, db, db2)
}

func TestUpdateDB_FactorySuccess(t *testing.T) {
	dbPath := t.TempDir() + "test-db"

	db, err := NewDefaultUpdateDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewUpdateDB(dbPath, nil, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewReadOnlyUpdateDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	err = db.Put([]byte{}, []byte{})
	assert.Error(t, leveldb.ErrReadOnly, err)
	assert.Nil(t, db.Close())
}

func TestUpdateDB_FactoryError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultBaseDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err = NewDefaultBaseDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewUpdateDB(dbPath, nil, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewReadOnlyUpdateDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestUpdateDB_MakeDefaultUpdateDBFromBaseDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultUpdateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db2 := MakeDefaultUpdateDBFromBaseDB(db)
	assert.NotNil(t, db2)
	assert.Equal(t, db, db2)
}

func TestCodeDB_FactorySuccess(t *testing.T) {
	dbPath := t.TempDir() + "test-db"

	db, err := NewDefaultCodeDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewCodeDB(dbPath, nil, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewReadOnlyCodeDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	err = db.Put([]byte{}, []byte{})
	assert.Error(t, leveldb.ErrReadOnly, err)
	assert.Nil(t, db.Close())
}

func TestCodeDB_FactoryError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultCodeDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err = NewDefaultCodeDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewCodeDB(dbPath, nil, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewReadOnlyCodeDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestCodeDB_MakeDefaultCodeDBFromBaseDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultCodeDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db2 := MakeDefaultCodeDBFromBaseDB(db)
	assert.NotNil(t, db2)
	assert.Equal(t, db, db2)
}

func TestDestroyedAccountDB_FactorySuccess(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultDestroyedAccountDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewReadOnlyDestroyedAccountDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	err = db.SetDestroyedAccounts(1, 1, nil, nil)
	assert.Error(t, leveldb.ErrReadOnly, err)
	assert.Nil(t, db.Close())
}

func TestDestroyedAccountDB_FactoryError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultDestroyedAccountDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err = NewDefaultDestroyedAccountDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewReadOnlyDestroyedAccountDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestSubstateDB_MakeDefaultDestroyedAccountDBFromBaseDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultDestroyedAccountDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db2 := MakeDefaultDestroyedAccountDBFromBaseDB(db.backend)
	assert.NotNil(t, db2)
	assert.Equal(t, db, db2)
}

func TestSubstateDB_FactorySuccess(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewSubstateDB(dbPath, nil, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())

	db, err = NewReadOnlySubstateDB(dbPath)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	err = db.Put([]byte{}, []byte{})
	assert.Error(t, leveldb.ErrReadOnly, err)
	assert.Nil(t, db.Close())
}

func TestSubstateDB_FactoryError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err = NewDefaultSubstateDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewSubstateDB(dbPath, nil, nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, db)

	db, err = NewReadOnlySubstateDB(dbPath)
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestSubstateDB_MakeDefaultSubstateDBFromBaseDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db2 := MakeDefaultSubstateDBFromBaseDB(db)
	assert.NotNil(t, db2)
}

func TestSubstateDB_MakeDefaultSubstateDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// cast interface to struct
	backend := db.getBackend().(*leveldb.DB)
	db2 := MakeDefaultSubstateDB(backend)
	assert.NotNil(t, db2)
}

func TestSubstateDB_MakeSubstateDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// cast interface to struct
	backend := db.getBackend().(*leveldb.DB)
	db2 := MakeSubstateDB(backend, nil, nil)
	assert.NotNil(t, db2)
}
