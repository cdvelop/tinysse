package tinysse

import "testing"

func TestNew(t *testing.T) {
	c := &Config{}
	sse := New(c)
	if sse == nil {
		t.Fatal("New() returned nil")
	}
	if sse.config != c {
		t.Error("New() did not set config")
	}
}
