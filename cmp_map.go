package diff

import "reflect"

func cmpMap(df *differ, steps []string, k, v reflect.Type, lv, rv reflect.Value) bool {
	if lv.Len() != rv.Len() {
		return df.Callback(buildPath(steps), DiffOfMapLength, lv, rv)
	}
	keys := lv.MapKeys()
	for _, key := range keys {
		lvv, rvv := lv.MapIndex(key), rv.MapIndex(key)
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
