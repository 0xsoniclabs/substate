// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rlp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type testEncoder struct {
	err error
}

func (e *testEncoder) EncodeRLP(w io.Writer) error {
	if e == nil {
		panic("EncodeRLP called on nil value")
	}
	if e.err != nil {
		return e.err
	}
	w.Write([]byte{0, 1, 0, 1, 0, 1, 0, 1, 0, 1})
	return nil
}

type testEncoderValueMethod struct{}

func (e testEncoderValueMethod) EncodeRLP(w io.Writer) error {
	w.Write([]byte{0xFA, 0xFE, 0xF0})
	return nil
}

type byteEncoder byte

var EmptyList = []byte{0xC0}

func (e byteEncoder) EncodeRLP(w io.Writer) error {
	w.Write(EmptyList)
	return nil
}

type undecodableEncoder func()

func (f undecodableEncoder) EncodeRLP(w io.Writer) error {
	w.Write([]byte{0xF5, 0xF5, 0xF5})
	return nil
}

type encodableReader struct {
	A, B uint
}

func (e *encodableReader) Read(b []byte) (int, error) {
	panic("called")
}

type namedByteType byte

var (
	_ = Encoder(&testEncoder{})
	_ = Encoder(byteEncoder(0))

	reader io.Reader = &encodableReader{1, 2}
)

type encTest struct {
	val           interface{}
	output, error string
}

var encTests = []encTest{
	// integers
	{val: uint32(0), output: "80"},
	{val: uint32(127), output: "7F"},
	{val: uint32(128), output: "8180"},
	{val: uint32(256), output: "820100"},
	{val: uint32(1024), output: "820400"},
	{val: uint32(0xFFFFFF), output: "83FFFFFF"},
	{val: uint32(0xFFFFFFFF), output: "84FFFFFFFF"},
	{val: uint64(0xFFFFFFFF), output: "84FFFFFFFF"},
	{val: uint64(0xFFFFFFFFFF), output: "85FFFFFFFFFF"},
	{val: uint64(0xFFFFFFFFFFFF), output: "86FFFFFFFFFFFF"},
	{val: uint64(0xFFFFFFFFFFFFFF), output: "87FFFFFFFFFFFFFF"},
	{val: uint64(0xFFFFFFFFFFFFFFFF), output: "88FFFFFFFFFFFFFFFF"},

	// byte arrays
	{val: [0]byte{}, output: "80"},
	{val: [1]byte{0}, output: "00"},
	{val: [1]byte{1}, output: "01"},
	{val: [1]byte{0x7F}, output: "7F"},
	{val: [1]byte{0x80}, output: "8180"},
	{val: [1]byte{0xFF}, output: "81FF"},
	{val: [3]byte{1, 2, 3}, output: "83010203"},
	{val: [57]byte{1, 2, 3}, output: "B839010203000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},

	// named byte type arrays
	{val: [0]namedByteType{}, output: "80"},
	{val: [1]namedByteType{0}, output: "00"},
	{val: [1]namedByteType{1}, output: "01"},
	{val: [1]namedByteType{0x7F}, output: "7F"},
	{val: [1]namedByteType{0x80}, output: "8180"},
	{val: [1]namedByteType{0xFF}, output: "81FF"},
	{val: [3]namedByteType{1, 2, 3}, output: "83010203"},
	{val: [57]namedByteType{1, 2, 3}, output: "B839010203000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},

	// slices
	{val: []uint{}, output: "C0"},
	{val: []uint{1, 2, 3}, output: "C3010203"},
	{
		// [ [], [[]], [ [], [[]] ] ]
		val:    []interface{}{[]interface{}{}, [][]interface{}{{}}, []interface{}{[]interface{}{}, [][]interface{}{{}}}},
		output: "C7C0C1C0C3C0C1C0",
	},
	{
		val:    []interface{}{uint(1), uint(0xFFFFFF), []interface{}{[]uint{4, 5, 5}}},
		output: "CA0183FFFFFFC4C3040505",
	},

	// Verify the error for unsupported type.
	{val: "string", error: "rlp: type string is not RLP-serializable"},
}

func runEncTests(t *testing.T, f func(val interface{}) ([]byte, error)) {
	for i, test := range encTests {
		output, err := f(test.val)
		if err != nil && test.error == "" {
			t.Errorf("test %d: unexpected error: %v\nvalue %#v\ntype %T",
				i, err, test.val, test.val)
			continue
		}
		if test.error != "" && fmt.Sprint(err) != test.error {
			t.Errorf("test %d: error mismatch\ngot   %v\nwant  %v\nvalue %#v\ntype  %T",
				i, err, test.error, test.val, test.val)
			continue
		}
		if err == nil && !bytes.Equal(output, unhex(test.output)) {
			t.Errorf("test %d: output mismatch:\ngot   %X\nwant  %s\nvalue %#v\ntype  %T",
				i, output, test.output, test.val, test.val)
		}
	}
}

func TestEncode(t *testing.T) {
	runEncTests(t, func(val interface{}) ([]byte, error) {
		b := new(bytes.Buffer)
		err := Encode(b, val)
		return b.Bytes(), err
	})
}

func TestEncodeToBytes(t *testing.T) {
	runEncTests(t, EncodeToBytes)
}

func TestEncodeToReader(t *testing.T) {
	runEncTests(t, func(val interface{}) ([]byte, error) {
		_, r, err := EncodeToReader(val)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(r)
	})
}

func TestEncodeToReaderPiecewise(t *testing.T) {
	runEncTests(t, func(val interface{}) ([]byte, error) {
		size, r, err := EncodeToReader(val)
		if err != nil {
			return nil, err
		}

		// read output piecewise
		output := make([]byte, size)
		for start, end := 0, 0; start < size; start = end {
			if remaining := size - start; remaining < 3 {
				end += remaining
			} else {
				end = start + 3
			}
			n, err := r.Read(output[start:end])
			end = start + n
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
		}
		return output, nil
	})
}

// This is a regression test verifying that encReader
// returns its encbuf to the pool only once.
func TestEncodeToReaderReturnToPool(t *testing.T) {
	buf := make([]byte, 50)
	wg := new(sync.WaitGroup)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				_, r, _ := EncodeToReader([]byte{1, 2, 3, 4, 5})
				ioutil.ReadAll(r)
				r.Read(buf)
				r.Read(buf)
				r.Read(buf)
				r.Read(buf)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

var sink interface{}

func BenchmarkIntsize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = intsize(0x12345678)
	}
}

func BenchmarkPutint(b *testing.B) {
	buf := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		putint(buf, 0x12345678)
		sink = buf
	}
}

func BenchmarkEncodeConcurrentInterface(b *testing.B) {
	type struct1 struct {
		A string
		B int
		C [20]byte
	}
	value := []interface{}{
		uint(999),
		&struct1{A: "hello", B: 0xFFFFFFFF},
		[10]byte{1, 2, 3, 4, 5, 6},
		[]string{"yeah", "yeah", "yeah"},
	}

	var wg sync.WaitGroup
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var buffer bytes.Buffer
			for i := 0; i < b.N; i++ {
				buffer.Reset()
				err := Encode(&buffer, value)
				if err != nil {
					panic(err)
				}
			}
		}()
	}
	wg.Wait()
}

type byteArrayStruct struct {
	A [20]byte
	B [32]byte
	C [32]byte
}

func BenchmarkEncodeByteArrayStruct(b *testing.B) {
	var out bytes.Buffer
	var value byteArrayStruct

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out.Reset()
		if err := Encode(&out, &value); err != nil {
			b.Fatal(err)
		}
	}
}

func unhex(str string) []byte {
	b, err := hex.DecodeString(strings.Replace(str, " ", "", -1))
	if err != nil {
		panic(fmt.Sprintf("invalid hex string: %q", str))
	}
	return b
}
