package main

import (
	"testing"

	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/utils"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCompareSubstate_Simple(t *testing.T) {
	src := t.TempDir() + "src-db"
	target := t.TempDir() + "target-db"

	srcDb, err := db.NewDefaultSubstateDB(src)
	if err != nil {
		t.Fatal(err)
	}
	if srcDb == nil {
		t.Fatal("srcDb is nil")
	}
	err = srcDb.Close()
	if err != nil {
		t.Fatal(err)
	}

	targetDb, err := db.NewDefaultSubstateDB(target)
	if err != nil {
		t.Fatal(err)
	}
	if targetDb == nil {
		t.Fatal("targetDb is nil")
	}
	err = targetDb.Close()
	if err != nil {
		t.Fatal(err)
	}

	args := []string{
		"dummy",
		"--workers", "1",
		"--src", src,
		"--target", target,
		"--block-segment", "0_1000000",
	}
	app := &cli.App{
		Name:   "test",
		Action: compare,
		Flags: []cli.Flag{
			&utils.WorkersFlag,
			&utils.SrcDbFlag,
			&utils.TargetDbFlag,
			&utils.BlockSegmentFlag,
		},
	}
	err = app.Run(args)
	assert.NoError(t, err)
}

func TestCompareSubstate_SrcError(t *testing.T) {
	src := t.TempDir() + "src-db"
	target := t.TempDir() + "target-db"

	args := []string{
		"dummy",
		"--workers", "1",
		"--src", src,
		"--target", target,
		"--block-segment", "0_1000000",
	}
	app := &cli.App{
		Name:   "test",
		Action: compare,
		Flags: []cli.Flag{
			&utils.WorkersFlag,
			&utils.SrcDbFlag,
			&utils.TargetDbFlag,
			&utils.BlockSegmentFlag,
		},
	}
	err := app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot open leveldb")
}

func TestCompareSubstate_TargetError(t *testing.T) {
	src := t.TempDir() + "src-db"
	target := t.TempDir() + "target-db"

	srcDb, err := db.NewDefaultSubstateDB(src)
	if err != nil {
		t.Fatal(err)
	}
	if srcDb == nil {
		t.Fatal("srcDb is nil")
	}
	err = srcDb.Close()
	if err != nil {
		t.Fatal(err)
	}

	targetDb, err := db.NewDefaultSubstateDB(target)
	if err != nil {
		t.Fatal(err)
	}
	if targetDb == nil {
		t.Fatal("targetDb is nil")
	}
	defer func(targetDb db.SubstateDB) {
		err := targetDb.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(targetDb)

	args := []string{
		"dummy",
		"--workers", "1",
		"--src", src,
		"--target", target,
		"--block-segment", "0_1000000",
	}
	app := &cli.App{
		Name:   "test",
		Action: compare,
		Flags: []cli.Flag{
			&utils.WorkersFlag,
			&utils.SrcDbFlag,
			&utils.TargetDbFlag,
			&utils.BlockSegmentFlag,
		},
	}
	err = app.Run(args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot open leveldb")
}
