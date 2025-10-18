package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestUpdateDB_MakeDefaultUpdateDBFromBaseDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultUpdateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db2, err := MakeDefaultUpdateDBFromBaseDB(db)
	assert.NoError(t, err)
	assert.NotNil(t, db2)
	assert.Equal(t, db.GetSubstateEncoding(), db2.GetSubstateEncoding())
	assert.Equal(t, db2.GetBackend(), db.GetBackend())
}

func TestUpdateDB_MakeDefaultUpdateDBFromBaseDBWithEncoding(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultUpdateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		db2, err := MakeDefaultUpdateDBFromBaseDBWithEncoding(db, DefaultEncodingSchema)
		assert.NoError(t, err)
		assert.NotNil(t, db2)
		assert.Equal(t, db.GetSubstateEncoding(), db2.GetSubstateEncoding())
		assert.Equal(t, db2.GetBackend(), db.GetBackend())
	})

	t.Run("error", func(t *testing.T) {
		db2, err := MakeDefaultUpdateDBFromBaseDBWithEncoding(db, "invalid-schema")
		assert.Error(t, err)
		assert.Nil(t, db2)
	})
}

func TestCodeDB_ConstructorSuccess(t *testing.T) {
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

func TestCodeDB_ConstructorError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	_, err := NewDefaultCodeDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewDefaultCodeDB(dbPath)
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

func TestDestroyedAccountDB_ConstructorSuccess(t *testing.T) {
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

func TestDestroyedAccountDB_ConstructorError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	_, err := NewDefaultDestroyedAccountDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewDefaultDestroyedAccountDB(dbPath)
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

	db2, err := MakeDefaultDestroyedAccountDBFromBaseDB(db)
	assert.NoError(t, err)
	assert.NotNil(t, db2)
	assert.Equal(t, db2.GetSubstateEncoding(), db.GetSubstateEncoding())
	assert.Equal(t, db2.GetBackend(), db.GetBackend())
}

func TestSubstateDB_MakeDefaultDestroyedAccountDBFromBaseDBWithEncoding(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultDestroyedAccountDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		db2, err := MakeDefaultDestroyedAccountDBFromBaseDBWithEncoding(db, DefaultEncodingSchema)
		assert.NoError(t, err)
		assert.NotNil(t, db2)
		assert.Equal(t, db2.GetSubstateEncoding(), db.GetSubstateEncoding())
		assert.Equal(t, db2.GetBackend(), db.GetBackend())
	})

	t.Run("error", func(t *testing.T) {
		db2, err := MakeDefaultDestroyedAccountDBFromBaseDBWithEncoding(db, "invalid-schema")
		assert.Error(t, err)
		assert.Nil(t, db2)
	})
}

func TestSubstateDB_ConstructorSuccess(t *testing.T) {
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

func TestSubstateDB_ConstructorError(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	_, err := NewDefaultSubstateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewDefaultSubstateDB(dbPath)
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

	db2, err := MakeDefaultSubstateDBFromBaseDB(db)
	assert.NoError(t, err)
	assert.NotNil(t, db2)
}

func TestSubstateDB_MakeDefaultSubstateDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	assert.NoError(t, err)

	// cast interface to struct
	backend := db.GetBackend().(*leveldb.DB)
	db2, err := MakeDefaultSubstateDB(backend)
	assert.NoError(t, err)
	assert.NotNil(t, db2)
}

func TestSubstateDB_MakeSubstateDB(t *testing.T) {
	dbPath := t.TempDir() + "test-db"
	db, err := NewDefaultSubstateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// cast interface to struct
	backend := db.GetBackend().(*leveldb.DB)
	db2, err := MakeSubstateDB(backend, nil, nil)
	assert.NotNil(t, db2)
	assert.NoError(t, err)
}
