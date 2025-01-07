package db

import (
	"strings"
	"testing"

	pb "github.com/0xsoniclabs/substate/protobuf"
	"github.com/0xsoniclabs/substate/rlp"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
)

type encTest struct {
	bytes []byte
	blk   uint64
	tx    int
}

var (
	blk = testSubstate.Block
	tx  = testSubstate.Transaction

	simplePb, _ = pb.Encode(testSubstate, blk, tx)
	testPb      = encTest{bytes: simplePb, blk: blk, tx: tx}

	simpleRlp, _ = trlp.EncodeToBytes(rlp.NewRLP(testSubstate))
	testRlp      = encTest{bytes: simpleRlp, blk: blk, tx: tx}

	supportedEncoding = map[string]encTest{
		"rlp":      testRlp,
		"protobuf": testPb,
	}
)

func TestSubstateEncoding_NilEncodingDefaultsToRlp(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := newSubstateDB(path, nil, nil, nil)
	if err != nil {
		t.Errorf("cannot open db; %v", err)
	}

	// purposely never set encoding

	// defaults to rlp
	if got := db.GetSubstateEncoding(); got != "rlp" {
		t.Fatalf("substate encoding should be nil, got: %s", got)
	}

	_, err = db.decodeToSubstate(testRlp.bytes, testRlp.blk, testRlp.tx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubstateEncoding_DefaultEncodingDefaultsToRlp(t *testing.T) {
	defaultKeywords := []string{"", "default"}
	for _, defaultEncoding := range defaultKeywords {
		path := t.TempDir() + "test-db-" + defaultEncoding
		db, err := newSubstateDB(path, nil, nil, nil)
		if err != nil {
			t.Errorf("cannot open db; %v", err)
		}

		_, err = db.SetSubstateEncoding(defaultEncoding)
		if err != nil {
			t.Fatalf("Default encoding '%s' must be supported, but error", defaultEncoding)
		}

		_, err = db.decodeToSubstate(testRlp.bytes, testRlp.blk, testRlp.tx)
		if err != nil {
			t.Fatal(err)
		}

		if got := db.GetSubstateEncoding(); got != "rlp" {
			t.Fatalf("db should default to rlp, got: %s", got)
		}
	}
}

func TestSubstateEncoding_UnsupportedEncodingThrowsError(t *testing.T) {
	path := t.TempDir() + "test-db"
	db, err := newSubstateDB(path, nil, nil, nil)
	if err != nil {
		t.Errorf("cannot open db; %v", err)
	}

	_, err = db.SetSubstateEncoding("EncodingNotSupported")
	if err == nil || !strings.Contains(err.Error(), "encoding not supported") {
		t.Error("encoding not supported, but no error")
	}
}

func TestSubstateEncoding_TestDb(t *testing.T) {
	for encoding, et := range supportedEncoding {
		path := t.TempDir() + "test-db-" + encoding
		db, err := newSubstateDB(path, nil, nil, nil)
		if err != nil {
			t.Errorf("cannot open db; %v", err)
		}

		db, err = db.SetSubstateEncoding(encoding)
		if err != nil {
			t.Error(err)
		}

		ss, err := db.decodeToSubstate(et.bytes, et.blk, et.tx)
		if err != nil {
			t.Error(err)
		}

		err = addCustomSubstate(db, et.blk, ss)
		if err != nil {
			t.Error(err)
		}

		testSubstateDB_GetSubstate(db, t)
	}
}

func TestSubstateEncoding_TestIterator(t *testing.T) {
	for encoding, et := range supportedEncoding {
		path := t.TempDir() + "test-db-" + encoding
		db, err := newSubstateDB(path, nil, nil, nil)
		if err != nil {
			t.Errorf("cannot open db; %v", err)
		}

		_, err = db.SetSubstateEncoding(encoding)
		if err != nil {
			t.Error(err)
		}

		ss, err := db.decodeToSubstate(et.bytes, et.blk, et.tx)
		if err != nil {
			t.Error(err)
		}

		err = addCustomSubstate(db, et.blk, ss)
		if err != nil {
			t.Error(err)
		}

		testSubstatorIterator_Value(db, t)
	}
}
