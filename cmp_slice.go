package diff

import (
	"reflect"
	"sort"
)

type sliceElem struct {
	idx      int
	identity string
}
type sliceElems []sliceElem

func (a sliceElems) Len() int      { return len(a) }
func (a sliceElems) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sliceElems) Less(i, j int) bool {
	if a[i].identity == a[j].identity {
		return a[i].idx < a[j].idx
	}
	return a[i].identity < a[j].identity
}

type getIDFunc func(reflect.Value) string

func buildGetIDFn(df *differ, et reflect.Type) getIDFunc {
	if et == nil {
		return func(reflect.Value) string { return _ZERO }
	}
	if fn, ok := df.typeIDFuncs[et]; ok {
		return func(_v reflect.Value) string {
			if !_v.IsValid() {
				return _ZERO
			}
			out := fn.Call([]reflect.Value{_v})
			return out[0].String()
		}
	}
	var kind reflect.Kind
	if et.Kind() == reflect.Ptr {
		kind = et.Elem().Kind()
	} else {
		kind = et.Kind()
	}
	fn, ok := df.kindIDFuncs[kind]
	if !ok {
		fn = reflect.ValueOf(df.IDOfAnything)
	}
	if et.Kind() == reflect.Ptr {
		return func(_v reflect.Value) string {
			if !_v.IsValid() || !_v.Elem().IsValid() {
				return _ZERO
			}
			out := fn.Call([]reflect.Value{_v.Elem().Convert(fn.Type().In(0))})
			return out[0].String()
		}
	}
	return func(_v reflect.Value) string {
		if !_v.IsValid() {
			return _ZERO
		}
		out := fn.Call([]reflect.Value{_v.Convert(fn.Type().In(0))})
		return out[0].String()
	}
}

func buildSliceElems(df *differ, et reflect.Type, v reflect.Value, fn getIDFunc) sliceElems {
	ids := make(sliceElems, v.Len())
	for i := 0; i < v.Len(); i++ {
		ids[i] = sliceElem{
			identity: fn(v.Index(i)),
			idx:      i,
		}
	}
	sort.Sort(ids)
	return ids
}

func alignSlice(ls, rs sliceElems) (left sliceElems, right sliceElems, added sliceElems, deleted sliceElems) {
	var iL, iR int
	for iL < len(ls) && iR < len(rs) {
		switch {
		case ls[iL].identity < rs[iR].identity:
			deleted = append(deleted, ls[iL])
			iL++
		case ls[iL].identity > rs[iR].identity:
			added = append(added, rs[iR])
			iR++
		default:
			// equal
			left = append(left, ls[iL])
			right = append(right, rs[iR])
			iL++
			iR++
		}
	}
	if iL < len(ls) {
		deleted = append(deleted, ls[iL:]...)
	}
	if iR < len(rs) {
		added = append(added, rs[iR:]...)
	}

	if len(deleted) > 0 && len(added) > 0 {
		if len(deleted) > len(added) {
			left = append(left, deleted[:len(added)]...)
			right = append(right, added...)
			deleted = deleted[len(added):]
			added = nil
		} else {
			left = append(left, deleted...)
			right = append(right, added[:len(deleted)]...)
			added = added[len(deleted):]
			deleted = nil
		}
	}

	return
}

func cmpSlice(df *differ, steps []string, et reflect.Type, lv, rv reflect.Value) bool {
	df.setPathToType(buildPath(appendPath(steps, buildIndexStep(0))), et)

	getIDFn := buildGetIDFn(df, et)
	var addedID, deletedID sliceElems
	leftID := buildSliceElems(df, et, lv, getIDFn)
	rightID := buildSliceElems(df, et, rv, getIDFn)
	leftID, rightID, addedID, deletedID = alignSlice(leftID, rightID)
	for _, elem := range deletedID {
		step := buildIndexStep(elem.idx)
		if !df.Callback(buildPath(appendPath(steps, step)), DiffOfLeftElemRemoved, lv.Index(elem.idx), defaultValue(et)) {
			return false
		}
	}
	for _, elem := range addedID {
		step := buildIndexStep(elem.idx)
		if !df.Callback(buildPath(appendPath(steps, step)), DiffOfRightElemAdded, defaultValue(et), rv.Index(elem.idx)) {
			return false
		}
	}
	if et.Kind() == reflect.Ptr {
		for i, lelem := range leftID {
			relem := rightID[i]
			lvv, rvv := lv.Index(lelem.idx), rv.Index(relem.idx)
			if lvv.IsNil() != rvv.IsNil() {
				if lvv.IsNil() && !rvv.IsNil() {
					p := buildPath(appendPath(steps, buildIndexStep(lelem.idx)))
					if !df.Callback(p, DiffOfLeftNoValue, lvv, rvv) {
						return false
					}
				} else if !lvv.IsNil() && rvv.IsNil() {
					p := buildPath(appendPath(steps, buildIndexStep(lelem.idx)))
					if !df.Callback(p, DiffOfRightNoValue, lvv, rvv) {
						return false
					}
				}
				continue
			}
			if !lvv.IsNil() {
				if !cmpVal(df, appendPath(steps, buildIndexStep(lelem.idx)), et.Elem(), lvv.Elem(), rvv.Elem()) {
					return false
				}
			}
		}
	} else {
		for i, lelem := range leftID {
			relem := rightID[i]
			lvv, rvv := lv.Index(lelem.idx), rv.Index(relem.idx)
			if !cmpVal(df, appendPath(steps, buildIndexStep(lelem.idx)), et, lvv, rvv) {
				return false
			}
		}
	}
	return true
}
