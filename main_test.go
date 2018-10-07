package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	var e string
	// ---------
	e = Getenv("YOLO", "nope")
	assert.Equal(t, e, "nope")
	// ----------
	os.Setenv("YOLO", "woohoo")
	e = Getenv("YOLO", "nope")
	assert.Equal(t, e, "woohoo")
}

func TestGoroutines(t *testing.T) {
	var r interface{}
	r = Goroutines()

	switch r := r.(type) {
	case int:
		t.Log(int(r))
	default:
		t.Fatal("wrong type")
	}
}

func TestUptime(t *testing.T) {
	var u interface{}
	u = Uptime()
	switch u := u.(type) {
	case int64:
		t.Log(int64(u))
	default:
		t.Fatal("wrong type")
	}
}
