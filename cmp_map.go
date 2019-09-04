package diff

import (
	"reflect"
)

func cmpMap(df *differ, steps []string, k, v reflect.Type, lv, rv reflect.Value) bool {
	visitedKeys := make(map[interface{}]bool)
	keys := lv.MapKeys()
	for _, key := range keys {
		visitedKeys[key.Interface()] = true
		lvv, rvv := lv.MapIndex(key), rv.MapIndex(key)
		df.setPathToType(buildPath(appendPath(steps, key.String())), lvv.Type())
		if !rvv.IsValid() {
			if !df.Callback(buildPath(appendPath(steps, key.String())), DiffOfRightNoValue, lvv, rvv) {
				return false
			}
			continue
		}
		if lvv.Type().Kind() == reflect.Ptr {
			if lvv.IsNil() != rvv.IsNil() {
				s := appendPath(steps, key.String())
				if lvv.IsNil() && !rvv.IsNil() {
					if !df.Callback(buildPath(s), DiffOfLeftNoValue, lvv, rvv) {
						return false
					}
				} else if !lvv.IsNil() && rvv.IsNil() {
					if !df.Callback(buildPath(s), DiffOfRightNoValue, lvv, rvv) {
						return false
					}
				}
				continue
			}
			if !lvv.IsNil() {
				if !cmpVal(df, appendPath(steps, key.String()), lvv.Type().Elem(), lvv.Elem(), rvv.Elem()) {
					return false
				}
			}
		} else {
			if !cmpVal(df, appendPath(steps, key.String()), lvv.Type(), lvv, rvv) {
				return false
			}
		}

	}
	keys = rv.MapKeys()
	for _, key := range keys {
		if visitedKeys[key.Interface()] {
			continue
		}
		lvv, rvv := lv.MapIndex(key), rv.MapIndex(key)
		df.setPathToType(buildPath(appendPath(steps, key.String())), rvv.Type())
		if !lvv.IsValid() {
			if !df.Callback(buildPath(appendPath(steps, key.String())), DiffOfLeftNoValue, lvv, rvv) {
				return false
			}
			continue
		}
		if lvv.Type().Kind() == reflect.Ptr {
			if lvv.IsNil() != rvv.IsNil() {
				s := appendPath(steps, key.String())
				if lvv.IsNil() && !rvv.IsNil() {
					if !df.Callback(buildPath(s), DiffOfLeftNoValue, lvv, rvv) {
						return false
					}
				} else if !lvv.IsNil() && rvv.IsNil() {
					if !df.Callback(buildPath(s), DiffOfRightNoValue, lvv, rvv) {
						return false
					}
				}
				continue
			}
			if !lvv.IsNil() {
				if !cmpVal(df, appendPath(steps, key.String()), lvv.Type().Elem(), lvv.Elem(), rvv.Elem()) {
					return false
				}
			}
		} else {
			if !cmpVal(df, appendPath(steps, key.String()), lvv.Type(), lvv, rvv) {
				return false
			}
		}

	}
	return true
}
