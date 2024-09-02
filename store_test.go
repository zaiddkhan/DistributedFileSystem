package main

import (
	"bytes"
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
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "specials"

	data := []byte("hello world")
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
}
