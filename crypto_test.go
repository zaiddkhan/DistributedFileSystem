package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {

	payload := "hello world"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dst.String())

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Fatal(err)
	}
	if nw != 16+len(payload) {
		t.Fail()
	}
	if out.String() != payload {
		t.Errorf("Decryption failed")
	}
	fmt.Println(out.String())
}
