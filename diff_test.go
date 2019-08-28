package diff

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBasicArray(t *testing.T) {
	r1 := []string{"a", "b", "c", "e", "f"}
	r2 := []string{"a", "g", "b", "c", "a", "t"}
	df := New()
	var cnt int
	fn := func(p string, r Reason, lv, rv reflect.Value) bool {
		switch r {
		case DiffOfRightElemAdded:
			if rv.String() != "t" {
				t.Fatal("bad cmp")
			}
		case DiffOfValue:
			if rv.String() != "a" && rv.String() != "g" {
				t.Fatal("bad cmp")
			}
		}
		cnt++
		t.Log(p, r, lv.String(), rv.String())
		return true
	}
	if df.Compare(r1, r2, fn) {
		t.Fatal("should not equal")
	}
	if cnt != 3 {
		t.Fatal("bad cmp", cnt)
	}
}

func TestKind(t *testing.T) {
	df := New()
	var reason Reason
	df.Compare(1, 12, func(p string, r Reason, lv, rv reflect.Value) bool {
		reason = r
		return true
	})
	if reason != DiffOfValue {
		t.Fatal("should diff value")
	}
	df.Compare(1, int64(1), func(p string, r Reason, lv, rv reflect.Value) bool {
		reason = r
		return true
	})
	if reason != DiffOfType {
		t.Fatal("should diff value,but ", reason)
	}
	if !df.Compare("a", "a", nil) {
		t.Fatal("should equal")
	}
	if !df.Compare(3.0/7.0, 3.0/7.00, nil) {
		t.Fatal("should equal")
	}
	if !df.Compare(int64(1), int64(1), nil) {
		t.Fatal("should equal")
	}
	if !df.Compare(uint64(1), uint64(1), nil) {
		t.Fatal("should equal")
	}
	var s *string
	if df.Compare(stringPtr("s"), s, func(p string, r Reason, lv, rv reflect.Value) bool {
		if r != DiffOfRightNoValue {
			t.Fatal("bad reason")
		}
		return true
	}) {
		t.Fatal("should not equal")
	}
	var i *int
	if df.Compare(i, intPtr(1), func(p string, r Reason, lv, rv reflect.Value) bool {
		if r != DiffOfLeftNoValue {
			t.Fatal("bad reason", r)
		}
		return true
	}) {
		t.Fatal("should not equal")
	}
}

func TestDiffStruct1(t *testing.T) {
	type Inner struct {
		Num int
	}
	type Simple struct {
		String    string
		StringPtr *string
		Int       int
		IntPtr    *int
		Tm        *time.Time
		Inner     *Inner
	}
	s1 := Simple{
		String:    "love",
		StringPtr: stringPtr("love-ptr"),
		Int:       1,
		IntPtr:    intPtr(100),
		Inner:     &Inner{},
	}
	s2 := Simple{
		String:    "love",
		StringPtr: stringPtr("love-ptr"),
		Int:       1,
		IntPtr:    intPtr(100),
		Inner:     &Inner{},
	}
	if !CompareValue(s1, s2) {
		t.Fatal("should equal")
	}
	now := time.Now()
	s1.Tm = &now
	if CompareValue(s1, s2) {
		t.Fatal("should not equal")
	}
	s2.Tm = &now

	s1.Inner.Num = 100
	s2.Inner.Num = 99

	if CompareValue(s1, s2) {
		t.Fatal("should not equal")
	}
	df := New()
	err := df.RegistCompareFunc(func(i1, i2 *Inner) bool {
		if i1.Num > i2.Num {
			return i1.Num-i2.Num < 5
		}
		return i2.Num-i1.Num < 5
	})
	if err != nil {
		t.Fatal("should regist Ok", err)
	}
	if !df.Compare(s1, s2, func(p string, re Reason, l, r reflect.Value) bool {
		return true
	}) {
		t.Fatal("should equal")
	}
}

func TestDiffSlice(t *testing.T) {
	arr, brr := []string{"a", "c", "a", "b"}, []string{"d", "e", "a", "a", "b"}
	df := New()
	if df.Compare(arr, brr, func(p string, re Reason, l, r reflect.Value) bool {
		if re == DiffOfRightElemAdded {
			if str := r.String(); str != "d" && str != "e" {
				t.Fatal("should add d/e")
			}
			if r.String() == "d" && p != ".[0]" {
				t.Fatal("bad new path")
			}
			if r.String() == "e" && p != ".[1]" {
				t.Fatal("bad new path")
			}
		} else if re == DiffOfLeftElemRemoved {
			if l.String() != "c" || p != ".[1]" {
				t.Fatal("bad new path", p)
			}
		}
		return true
	}) {
		t.Fatal("should not equal")
	}
}

func TestDiffComplexSlice(t *testing.T) {
	type Obj struct {
		ID   string
		Name string
		Age  int
	}
	idfn := func(obj Obj) string { return obj.ID }
	df := New()
	if err := df.RegistIDFunc(idfn); err != nil {
		t.Fatal("should pass")
	}
	arr := []Obj{
		{ID: "1", Name: "hello1"},
		{ID: "2", Name: "hello2-0"},
		{ID: "2", Name: "hello2-1"},
		{ID: "3", Name: "hello3"},
	}
	brr := []Obj{
		{ID: "1", Name: "hello100", Age: 12},
		{ID: "2", Name: "hello2-0"},
		{ID: "4", Name: "hello4"},
	}
	if df.Compare(arr, brr, func(p string, re Reason, l, r reflect.Value) bool {
		switch re {
		case DiffOfValue:
			if p == ".[0].Name" {
				lobj, robj := l.String(), r.String()
				if lobj != "hello1" && robj != "hello100" {
					t.Fatal("bad cmp")
				}
			} else if p == ".[0].Age" {
				if r.Int() != 12 {
					t.Fatal("bad cmp")
				}
			}
		case DiffOfLeftElemRemoved:
			obj := l.Interface().(Obj)
			if obj.Name != "hello2-1" && obj.Name != "hello3" {
				t.Fatal("bad cmp")
			}
		case DiffOfRightElemAdded:
			obj := r.Interface().(Obj)
			if obj.Name != "hello4" {
				t.Fatal("bad cmp")
			}
		}
		return true
	}) {
		t.Fatal("should not equal")
	}
}

func TestOmitPath(t *testing.T) {
	type OmitStruct struct {
		Name string
		Age  int
	}
	o1 := &OmitStruct{Name: "n1", Age: 1}
	o2 := &OmitStruct{Name: "n2", Age: 1}
	df := New()
	if df.Compare(o1, o2, nil) {
		t.Fatal("should not equal")
	}
	df.OmitPath(".Name")
	if !df.Compare(o1, o2, nil) {
		t.Fatal("should equal")
	}
}

func TestUseCmpTypeFunc(t *testing.T) {
	type Inner struct {
		Num int
	}
	type Structx struct {
		Inner *Inner
	}
	s1 := &Structx{Inner: &Inner{Num: 100}}
	s2 := &Structx{Inner: &Inner{Num: 101}}
	df := New()
	if df.Compare(s1, s2, nil) {
		t.Fatal("should not equal")
	}
	df.RegistCompareFunc(func(l, r *Inner) bool {
		return l.Num/10 == r.Num/10
	})
	if !df.Compare(s1, s2, nil) {
		t.Fatal("should  equal")
	}
}

func TestUseIDFunc(t *testing.T) {
	type Inner struct {
		Num int
	}
	type Structx struct {
		Inner []*Inner
	}
	s1 := &Structx{Inner: []*Inner{&Inner{Num: 100}}}
	s2 := &Structx{Inner: []*Inner{&Inner{Num: 101}}}
	df := New()
	if df.Compare(s1, s2, nil) {
		t.Fatal("should not equal")
	}
	if df.Compare(s1, s2, func(p string, r Reason, lv, rv reflect.Value) bool {
		if r != DiffOfValue {
			t.Fatal("bad reason", r, lv.Interface(), rv.Interface())
		}
		return true
	}) {
		t.Fatal("should not  equal")
	}
	df.RegistIDFunc(func(x *Inner) string {
		return fmt.Sprint(x.Num / 10)
	})
	if df.Compare(s1, s2, func(p string, r Reason, lv, rv reflect.Value) bool {
		if r != DiffOfValue {
			t.Fatal("bad reason")
		}
		if p != ".Inner[0].Num" {
			t.Fatal("bad path", p)
		}
		if int(lv.Int()) != 100 {
			t.Fatal("bad v")
		}
		if int(rv.Int()) != 101 {
			t.Fatal("bad v")
		}
		return true
	}) {
		t.Fatal("should not  equal")
	}
}

type BasicInfo struct {
	Name   *string
	Mobile *string
	Email  *string
}
type Company struct {
	Name string
	Link *string
}
type School struct {
	ID   string
	Name string
	Area *int
}
type Nothing struct{ X int }
type CType int64

type CType2 int

const (
	ct1 CType = iota
	ct11
	ct111
)

const (
	ct2 CType2 = iota
	ct22
	ct222
)

type HugeStruct struct {
	ID          *string
	Name        string
	BasicInfo   BasicInfo
	CompanyList []*Company
	SchoolList  []School
	Tags        []string
	Numbers     []int64
	CType       CType
	CType2      *CType2
	Nothing     *Nothing
}

func TestComplexStruct(t *testing.T) {
	h1, h2 := makeHugeStruct(), makeHugeStruct()
	if !CompareValue(h1, h2) {
		t.Fatal("should equals")
	}

	df := New()
	df.RegistIDFunc(func(c *Company) string {
		return *c.Link
	})
	h2.CompanyList[1].Name = "AWS"
	fn := func(path string, r Reason, lv, rv reflect.Value) bool {
		if !strings.Contains(path, ".CompanyList") {
			t.Fatal("companay should not equal")
		}
		if r != DiffOfValue {
			t.Fatal("now companay should not equal by name")
		}
		if lv.String() != "aws" || rv.String() != "AWS" || path != ".CompanyList[1].Name" {
			t.Fatal("should diff name")
		}
		return true
	}
	if df.Compare(h1, h2, fn) {
		t.Fatal("should not equals")
	}
	h1, h2 = makeHugeStruct(), makeHugeStruct()
	if !df.Compare(h1, h2, fn) {
		t.Fatal("should equals")
	}

	h2.Tags = []string{"b", "b", "c"}
	fn = func(path string, r Reason, lv, rv reflect.Value) bool {
		if !strings.Contains(path, ".Tags") {
			t.Fatal("tags should not equal")
		}
		return true
	}
	if df.Compare(h1, h2, fn) {
		t.Fatal("should not equal")
	}

	h1, h2 = makeHugeStruct(), makeHugeStruct()
	h2.SchoolList[1].Area = intPtr(100)
	if df.Compare(h1, h2, nil) {
		t.Fatal("should not equals")
	}
	df.RegistCompareFunc(func(l, r School) bool {
		return l.ID == r.ID && l.Name == r.Name
	})
	if !df.Compare(h1, h2, nil) {
		t.Fatal("should equals")
	}

}

func makeHugeStruct() *HugeStruct {
	_ct2 := ct2
	return &HugeStruct{
		ID:   stringPtr("id-v"),
		Name: "name",
		BasicInfo: BasicInfo{
			Name:   stringPtr("basic name"),
			Mobile: stringPtr("111111"),
		},
		CompanyList: []*Company{
			&Company{Name: "google", Link: stringPtr("www.google.com")},
			&Company{Name: "aws", Link: stringPtr("www.amazon.com")},
		},
		SchoolList: []School{
			{ID: "school1", Name: "s1"},
			{ID: "school2", Name: "s2", Area: intPtr(1)},
		},
		Tags:    []string{"a", "b", "c"},
		Numbers: []int64{1, 2, 3},
		CType:   ct1,
		CType2:  &_ct2,
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
