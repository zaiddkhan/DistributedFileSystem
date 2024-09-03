package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "picturees"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "94b07aae708dc367d6e022579ee94a45fc5d488f"
	expectedPathName := "94b07/aae70/8dc36/7d6e0/22579/ee94a/45fc5/d488f"

	if pathKey.Pathname != expectedPathName {
		t.Error("pathname error")
	}
	if pathKey.Filename != expectedOriginalKey {
		t.Error("original error", expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {

	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 10; i++ {

		key := fmt.Sprintf("foo_%d", i)

		data := []byte("some jpeg bytes")
		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		b, _ := ioutil.ReadAll(r)
		if string(b) != string(data) {
			t.Error("data error")
		}

		if ok := s.Has(key); !ok {
			t.Error("key", key, "not found")
		}

		err = s.Delete(key)
		if err != nil {
			fmt.Println(err)
		}
		if ok := s.Has(key); ok {
			t.Error("Expected to not find the key")
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, store *Store) {
	if err := store.Clear(); err != nil {
		t.Error(err)
	}
}
