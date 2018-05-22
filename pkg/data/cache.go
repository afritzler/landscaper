package data

import (
	"fmt"
	"sync"
)

type Cacher interface {
	GetAll() (Iterator, error)
	Get(key interface{}) (interface{}, error)
	Key(interface{}) interface{}
}

type Cache interface {
	Cacher
	Iterator() Iterator
	Entries() Iterator
	Reset()
}

type _Cache struct {
	lock     sync.Mutex
	cacher   Cacher
	entries  map[interface{}]interface{}
	complete bool
}

func NewCache(c Cacher) Cache {
	return &_Cache{cacher: c, entries: nil, complete: false}
}

func (this *_Cache) Key(e interface{}) interface{} {
	return this.cacher.Key(e)
}

func (this *_Cache) Reset() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.entries = nil
	this.complete = false
}

func (this *_Cache) GetAll() (Iterator, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.entries == nil || !this.complete {
		elems, err := this.cacher.GetAll()
		if err != nil {
			return nil, err
		}
		this.entries = map[interface{}]interface{}{}
		this.complete = true
		for elems.HasNext() {
			e := elems.Next()
			this.entries[this.cacher.Key(e)] = e
		}
	}
	return NewMappedIterator(newMapEntryIterator(this.entries), func(e interface{}) interface{} {
		return e.(MapEntry).Value
	}), nil
}

func (this *_Cache) Get(key interface{}) (interface{}, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	var p interface{} = nil
	if this.entries != nil {
		p = this.entries[key]
	}
	if p == nil && !this.complete {
		elem, err := this.cacher.Get(key)
		if err != nil {
			return nil, err
		}
		if this.entries == nil {
			this.entries = map[interface{}]interface{}{}
		}
		this.entries[key] = elem
		p = elem
	}
	if p == nil {
		return nil, fmt.Errorf("'%s' not found", key)
	}
	return p, nil
}

func (this *_Cache) Iterator() Iterator {
	this.lock.Lock()
	defer this.lock.Unlock()

	return newMapEntryIterator(this.entries)
}

func (this _Cache) Entries() Iterator {
	return NewMappedIterator(this.Iterator(), func(e interface{}) interface{} {
		return e.(MapEntry).Value
	})
}

func newMapEntryIterator(m map[interface{}]interface{}) ResettableIterator {
	entries := make([]interface{}, len(m))
	if m != nil {
		i := 0
		for k, v := range m {
			entries[i] = MapEntry{k, v}
			i++
		}
	}
	return NewSliceIterator(entries)
}
