package diff

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

func CmpBool(left, right bool) bool       { return left == right }
func CmpInt(left, right int) bool         { return left == right }
func CmpInt8(left, right int8) bool       { return left == right }
func CmpInt16(left, right int16) bool     { return left == right }
func CmpInt32(left, right int32) bool     { return left == right }
func CmpInt64(left, right int64) bool     { return left == right }
func CmpUint(left, right uint) bool       { return left == right }
func CmpUint8(left, right uint8) bool     { return left == right }
func CmpUint16(left, right uint16) bool   { return left == right }
func CmpUint32(left, right uint32) bool   { return left == right }
func CmpUint64(left, right uint64) bool   { return left == right }
func CmpUintptr(left, right uintptr) bool { return left == right }
func CmpFloat32(left, right float32) bool { return left == right }
func CmpFloat64(left, right float64) bool { return left == right }
func CmpString(left, right string) bool   { return left == right }
func CmpUnsafePointer(left, right unsafe.Pointer) bool {
	return uintptr(left) == uintptr(right)
}

func IDOfBool(b bool) string {
	if b {
		return "t"
	}
	return "f"
}

func IDOfInt(i int) string {
	return strconv.FormatInt(int64(i), 10)
}
func IDOfInt8(i int8) string {
	return strconv.FormatInt(int64(i), 10)
}
func IDOfInt16(i int16) string {
	return strconv.FormatInt(int64(i), 10)
}
func IDOfInt32(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}
func IDOfInt64(i int64) string {
	return strconv.FormatInt(int64(i), 10)
}
func IDOfUint(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}
func IDOfUint8(i uint8) string {
	return strconv.FormatUint(uint64(i), 10)
}
func IDOfUint16(i uint16) string {
	return strconv.FormatUint(uint64(i), 10)
}
func IDOfUint32(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}
func IDOfUint64(i uint64) string {
	return strconv.FormatUint(uint64(i), 10)
}
func IDOfUintptr(i uintptr) string {
	return fmt.Sprint(i)
}
func IDOfString(s string) string {
	return s
}
func IDOfFloat32(f float32) string {
	return strconv.FormatFloat(float64(f), 'f', 6, 64)
}
func IDOfFloat64(f float64) string {
	return strconv.FormatFloat(float64(f), 'f', 6, 64)
}

// IDOfAnything a is any non-pointer value
func (df *differ) IDOfAnything(a interface{}) string {
	v := reflect.ValueOf(a)
	if fn, ok := df.typeCache.Get(v.Type()); ok {
		out := fn.Call([]reflect.Value{v})
		return out[0].String()
	}
	if v.Type().Kind() == reflect.Struct {
		if fn, ok := isStructWithIDField(v.Type()); ok {
			df.typeCache.Set(v.Type(), fn)
			out := fn.Call([]reflect.Value{v})
			return out[0].String()
		}
	}
	return fmt.Sprintf("%v", a)
}

func isStructWithIDField(v reflect.Type) (fn reflect.Value, ok bool) {
	if v.Kind() != reflect.Struct {
		return
	}
	var idx int = -1
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Name == "ID" || v.Field(i).Name == "Id" {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	if v.Field(idx).Type.Kind() == reflect.Ptr {
		if !IsPrimitiveType(v.Field(idx).Type.Elem()) {
			return
		}
		fn = reflect.ValueOf(func(any interface{}) string {
			if any == nil {
				return _ZERO
			}
			vv := reflect.ValueOf(any)
			if !vv.IsValid() {
				return _ZERO
			}
			if vv.Field(idx).IsNil() || !vv.Field(idx).IsValid() {
				return _ZERO
			}
			return valueToString(vv.Field(idx).Elem())
		})
		ok = true
		return
	}
	if !IsPrimitiveType(v.Field(idx).Type) {
		return
	}
	fn = reflect.ValueOf(func(any interface{}) string {
		if any == nil {
			return _ZERO
		}
		vv := reflect.ValueOf(any)
		if !vv.IsValid() {
			return _ZERO
		}
		if !vv.Field(idx).IsValid() {
			return _ZERO
		}
		return valueToString(vv.Field(idx))
	})
	ok = true
	return
}

type typeIDCache struct {
	c map[reflect.Type]reflect.Value
	*sync.RWMutex
}

func newTypeIDCache() *typeIDCache {
	typeCache := &typeIDCache{c: make(map[reflect.Type]reflect.Value), RWMutex: new(sync.RWMutex)}
	return typeCache
}

func (c *typeIDCache) Get(t reflect.Type) (reflect.Value, bool) {
	c.RLock()
	v, ok := c.c[t]
	c.RUnlock()
	return v, ok
}

func (c *typeIDCache) Set(t reflect.Type, v reflect.Value) {
	c.Lock()
	c.c[t] = v
	c.Unlock()
}
