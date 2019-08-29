package diff

import (
	"reflect"
)

// CompareValue whether equal
func CompareValue(l, r interface{}) bool {
	df := New()
	return df.Compare(l, r, nil)
}

// Compare with callback
func (df *Differ) Compare(l interface{}, r interface{}, fn Callback) (equal bool) {
	if fn == nil {
		fn = func(*D) bool { return false }
	}
	lv, rv := reflect.ValueOf(l), reflect.ValueOf(r)
	lt, rt := lv.Type(), rv.Type()
	if lt != rt {
		fn(buildD(_ROOT, DiffOfType, lv, rv))
		return
	}
	_differ := newDiffer(df, fn)
	cmpVal(_differ, []string{}, lt, lv, rv)
	return !_differ.differenceExist
}

// MakePatch of l and  r
func (df *Differ) MakePatch(l interface{}, r interface{}) Patch {
	var patch Patch
	fn := func(_d *D) bool {
		patch.add(_d)
		return true
	}
	df.Compare(l, r, fn)
	return patch
}
