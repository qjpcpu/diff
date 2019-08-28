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
		fn = func(string, Reason, reflect.Value, reflect.Value) bool { return false }
	}
	lv, rv := reflect.ValueOf(l), reflect.ValueOf(r)
	lt, rt := lv.Type(), rv.Type()
	if lt != rt {
		fn(_ROOT, DiffOfType, lv, rv)
		return
	}
	_differ := newDiffer(df, fn)
	cmpVal(_differ, []string{}, lt, lv, rv)
	return !_differ.differenceExist
}
