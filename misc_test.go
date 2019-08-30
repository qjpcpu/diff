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
	if err := df.RegistPathCompareFunc("path", func(a, b int) bool { return false }); err != nil {
		t.Fatal("should not fail")
	}
	if err := df.RegistIDFunc(func(a interface{}) string { return "" }); err != nil {
		t.Fatal("should not fail")
	}
	if err := df.RegistKindIDFunc(func(int) string { return "" }); err != nil {
		t.Fatal("should not fail")
	}
}

func TestAlignSlice(t *testing.T) {
	ls := sliceElems{
		{idx: 0, identity: "1"},
		{idx: 1, identity: "2"},
		{idx: 2, identity: "3"},
		{idx: 3, identity: "4"},
	}
	lr := sliceElems{
		{idx: 0, identity: "10"},
		{idx: 3, identity: "40"},
		{idx: 1, identity: "90"},
	}
	l, r, a, d := alignSlice(ls, lr)
	if len(a) != 0 {
		t.Fatal("bad align")
	}
	if len(d) != 1 {
		t.Fatal("bad align")
	}
	if len(l) != 3 || len(r) != 3 {
		t.Fatal("bad align")
	}
	ls = sliceElems{
		{idx: 0, identity: "a"},
		{idx: 1, identity: "b"},
		{idx: 2, identity: "c"},
		{idx: 3, identity: "d"},
	}
	lr = sliceElems{
		{idx: 2, identity: "a"},
		{idx: 0, identity: "c"},
		{idx: 1, identity: "d"},
	}
	l, r, a, d = alignSlice(ls, lr)
	if len(a) != 0 {
		t.Fatal("bad align")
	}
	if len(d) != 1 || d[0].identity != "b" {
		t.Fatal("bad align")
	}
	if len(l) != 3 || len(r) != 3 {
		t.Fatal("bad align")
	}
	ls = sliceElems{
		{idx: 2, identity: "a"},
		{idx: 0, identity: "c"},
		{idx: 1, identity: "d"},
	}
	lr = sliceElems{
		{idx: 0, identity: "a"},
		{idx: 1, identity: "b"},
		{idx: 2, identity: "c"},
		{idx: 3, identity: "d"},
	}
	l, r, a, d = alignSlice(ls, lr)
	if len(d) != 0 {
		t.Fatal("bad align")
	}
	if len(a) != 1 || a[0].identity != "b" {
		t.Fatal("bad align")
	}
	if len(l) != 3 || len(r) != 3 {
		t.Fatal("bad align")
	}
}

func TestReplacePathStar(t *testing.T) {
	p := "A.B[1].C.D[12].E[1]"
	p = replaceSliceIndexToStar(p)
	if p != "A.B[*].C.D[*].E[*]" {
		t.Fatal("bad replace:", p)
	}
}
