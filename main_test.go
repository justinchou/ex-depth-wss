package main

import (
	"testing"
)

func TestDemo(t *testing.T) {
	b := true
	if b {
		t.Log("bool ok")
	} else {
		t.Error("bool failed")
	}

	var s string
	if s == "" {
		t.Log("string ok")
	} else {
		t.Error("string failed")
	}
}
