package rlp

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type writer func(reflect.Value, *encbuf) error

// typeinfo is an entry in the type cache.
type typeinfo struct {
	writer    writer
	writerErr error // error from makeWriter
}

func (i *typeinfo) generate(typ reflect.Type, tags tags) {
	i.writer, i.writerErr = makeWriter(typ, tags)
}

type typeCache struct {
	cur atomic.Value

	// This lock synchronizes writers.
	mu   sync.Mutex
	next map[typekey]*typeinfo
}

func (c *typeCache) generate(typ reflect.Type, tags tags) *typeinfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	cur := c.cur.Load().(map[typekey]*typeinfo)
	if info := cur[typekey{typ, tags}]; info != nil {
		return info
	}

	// Copy cur to next.
	c.next = make(map[typekey]*typeinfo, len(cur)+1)
	for k, v := range cur {
		c.next[k] = v
	}

	// Generate.
	info := c.infoWhileGenerating(typ, tags)

	// next -> cur
	c.cur.Store(c.next)
	c.next = nil
	return info
}

func (c *typeCache) infoWhileGenerating(typ reflect.Type, tags tags) *typeinfo {
	key := typekey{typ, tags}
	if info := c.next[key]; info != nil {
		return info
	}
	// Put a dummy value into the cache before generating.
	// If the generator tries to lookup itself, it will get
	// the dummy value and won't call itself recursively.
	info := new(typeinfo)
	c.next[key] = info
	info.generate(typ, tags)
	return info
}

func (c *typeCache) info(typ reflect.Type) *typeinfo {
	key := typekey{Type: typ}
	if info := c.cur.Load().(map[typekey]*typeinfo)[key]; info != nil {
		return info
	}

	// Not in the cache, need to generate info for this type.
	return c.generate(typ, tags{})
}

var theTC = newTypeCache()

func newTypeCache() *typeCache {
	c := new(typeCache)
	c.cur.Store(make(map[typekey]*typeinfo))
	return c
}

// typekey is the key of a type in typeCache. It includes the struct tags because
// they might generate a different decoder.
type typekey struct {
	reflect.Type
	tags
}

// tags represents struct tags.
type tags struct {
	// rlp:"nil" controls whether empty input results in a nil pointer.
	nilOK bool

	// rlp:"optional" allows for a field to be missing in the input list.
	// If this is set, all subsequent fields must also be optional.
	optional bool

	// rlp:"tail" controls whether this field swallows additional list elements. It can
	// only be set for the last field, which must be of slice type.
	tail bool

	// rlp:"-" ignores fields.
	ignored bool
}

func isUint(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}

func isByte(typ reflect.Type) bool {
	return typ.Kind() == reflect.Uint8 && !typ.Implements(encoderInterface)
}

func cachedWriter(typ reflect.Type) (writer, error) {
	info := theTC.info(typ)
	return info.writer, info.writerErr
}
