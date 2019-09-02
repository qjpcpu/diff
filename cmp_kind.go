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
	vtype := reflect.TypeOf(a)
	if vtype == nil {
		return _ZERO
	}
	v := reflect.ValueOf(a)
	if fn, ok := df.typeCache.Get(vtype); ok {
		out := fn.Call([]reflect.Value{v})
		return out[0].String()
	}
	if vtype.Kind() == reflect.Struct {
		if fn, ok := isStructWithIDField(vtype); ok {
			df.typeCache.Set(vtype, fn)
			out := fn.Call([]reflect.Value{v})
			return out[0].String()
		}
	}
	if IsPrimitiveType(vtype) || (vtype.Kind() == reflect.Ptr && IsPrimitiveType(vtype.Elem())) {
		return fmt.Sprintf("%v", a)
	}
	return _ZERO
}

type valueModifier func(reflect.Value) reflect.Value

var invalidValue = reflect.ValueOf("invalid")

func isSimpleIDFiled(field reflect.StructField) bool {
	if field.Name != "ID" && field.Name != "Id" {
		return false
	}
	if IsPrimitiveType(field.Type) {
		return true
	}
	return field.Type.Kind() == reflect.Ptr && IsPrimitiveType(field.Type.Elem())
}

func maybeAnoymousSmpleIDFiled(field reflect.StructField) bool {
	if !field.Anonymous {
		return false
	}
	return field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct)
}

func findIDField(m []valueModifier, t reflect.Type) ([]valueModifier, bool) {
	if t.Kind() != reflect.Struct {
		return nil, false
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if isSimpleIDFiled(f) {
			return appendValueModifier(m, i, f.Type.Kind() == reflect.Ptr), true
		} else if maybeAnoymousSmpleIDFiled(f) {
			var sub reflect.Type
			if f.Type.Kind() == reflect.Ptr {
				sub = f.Type.Elem()
			} else {
				sub = f.Type
			}
			if ms, ok := findIDField(appendValueModifier(m, i, f.Type.Kind() == reflect.Ptr), sub); ok {
				return ms, ok
			}
		}
	}
	return nil, false
}

func appendValueModifier(m []valueModifier, fieldIndex int, isPtr bool) []valueModifier {
	return append(m, func(v reflect.Value) reflect.Value {
		if v == invalidValue || !v.IsValid() {
			return invalidValue
		}
		if isPtr {
			if v.Field(fieldIndex).IsNil() || !v.Field(fieldIndex).IsValid() {
				return invalidValue
			}
			return v.Field(fieldIndex).Elem()
		}
		if !v.Field(fieldIndex).IsValid() {
			return invalidValue
		}
		return v.Field(fieldIndex)
	})
}

func isStructWithIDField(t reflect.Type) (fn reflect.Value, ok bool) {
	if t.Kind() != reflect.Struct {
		return
	}
	var m []valueModifier
	if m, ok = findIDField([]valueModifier{}, t); !ok {
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
		for _, f := range m {
			vv = f(vv)
			if vv == invalidValue || !vv.IsValid() {
				return _ZERO
			}
		}
		return valueToString(vv)
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
