package diff

import (
	"reflect"
)

func cmpStruct(df *differ, steps []string, t reflect.Type, lv, rv reflect.Value) bool {
	for i := 0; i < lv.NumField(); i++ {
		lfv, rfv := lv.Field(i), rv.Field(i)
		ft := t.Field(i)
		if !isExported(ft.Name) {
			continue
		}

		df.setPathToType(buildPath(appendPath(steps, ft.Name)), ft.Type)

		if df.canCmpType(buildPath(appendPath(steps, ft.Name)), ft.Type) {
			if !df.cmpByType(appendPath(steps, ft.Name), ft.Type, lfv, rfv) {
				return false
			}
			continue
		}
		if ft.Type.Kind() == reflect.Ptr {
			if lfv.IsNil() != rfv.IsNil() {
				if lfv.IsNil() && !rfv.IsNil() {
					if !df.Callback(buildPath(appendPath(steps, ft.Name)), DiffOfLeftNoValue, lfv, rfv) {
						return false
					}
				} else if !lfv.IsNil() && rfv.IsNil() {
					if !df.Callback(buildPath(appendPath(steps, ft.Name)), DiffOfRightNoValue, lfv, rfv) {
						return false
					}
				}
				continue
			}
			if !lfv.IsNil() {
				if !cmpVal(df, appendPath(steps, ft.Name), ft.Type.Elem(), lfv.Elem(), rfv.Elem()) {
					return false
				}
			}
		} else {
			if !cmpVal(df, appendPath(steps, ft.Name), ft.Type, lfv, rfv) {
				return false
			}
		}
	}
	return true
}
