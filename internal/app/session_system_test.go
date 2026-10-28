package app

import (
	"testing"
)

func fakeChange(t *testing.T, value *int, target int) *systemChange[int] {
	t.Helper()
	dir := t.TempDir()
	systemDir = func() string { return dir }
	t.Cleanup(func() { systemDir = func() string { return "" } })
	return &systemChange[int]{
		file:   "fake.json",
		read:   func() (int, error) { return *value, nil },
		target: func(cur int) (int, bool) { return target, cur != target },
		write:  func(v int) error { *value = v; return nil },
	}
}

func TestSystemChangeRestoresAfterCrash(t *testing.T) {
	value := 7
	c := fakeChange(t, &value, 0)
	if err := c.on(); err != nil {
		t.Fatal(err)
	}
	if value != 0 || !c.applied() {
		t.Fatalf("after on value=%d applied=%v", value, c.applied())
	}
	value = 0
	c.on()
	restarted := &systemChange[int]{file: c.file, read: c.read, target: c.target, write: c.write}
	if err := restarted.off(); err != nil {
		t.Fatal(err)
	}
	if value != 7 || restarted.applied() {
		t.Fatalf("after recover value=%d applied=%v", value, restarted.applied())
	}
}

func TestSystemChangeSkipsWhenAlreadyInTarget(t *testing.T) {
	value := 0
	c := fakeChange(t, &value, 0)
	if err := c.on(); err != nil {
		t.Fatal(err)
	}
	if c.applied() {
		t.Fatal("recorded a change that was not needed")
	}
	if err := c.off(); err != nil || value != 0 {
		t.Fatalf("off without record changed value to %d (%v)", value, err)
	}
}
