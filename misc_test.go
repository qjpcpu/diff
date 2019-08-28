package diff

import (
	"testing"
	"time"
)

func TestRegist(t *testing.T) {
	df := New()
	f1 := func(string, int) bool { return false }
	f2 := func(string) bool { return false }
	f3 := func(*time.Time, time.Time) bool { return false }
	f4 := func(*time.Time, *time.Time) bool { return false }
	if err := df.RegistCompareFunc(f1); err == nil {
		t.Fatal("should fail")
	}
	if err := df.RegistCompareFunc(f2); err == nil {
		t.Fatal("should fail")
	}
	if err := df.RegistCompareFunc(f3); err == nil {
		t.Fatal("should fail")
	}
	if err := df.RegistCompareFunc(f4); err != nil {
		t.Fatal("should not fail")
	}
}
