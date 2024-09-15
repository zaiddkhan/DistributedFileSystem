package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {

	src := bytes.NewReader([]byte("hello world"))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dst.Bytes())
}
