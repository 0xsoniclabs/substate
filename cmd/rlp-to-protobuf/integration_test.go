package main

import (
	"testing"

	"github.com/0xsoniclabs/substate/db"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestRunRlpToProtobuf_Simple(t *testing.T) {
	src := t.TempDir() + "src-db"
	dst := t.TempDir() + "dst-db"

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

	args := []string{
		"dummy",
		"--workers", "1",
		"--src", src,
		"--dst", dst,
		"--block-segment", "0_1000000",
	}
	app := &cli.App{
		Name:   "test",
		Action: RunRlpToProtobuf,
		Flags: []cli.Flag{
			&WorkersFlag,
			&SrcDbFlag,
			&DstDbFlag,
			&SkipTransferTxsFlag,
			&SkipCallTxsFlag,
			&SkipCreateTxsFlag,
			&BlockSegmentFlag,
		},
	}
	err = app.Run(args)
	assert.NoError(t, err)
}
