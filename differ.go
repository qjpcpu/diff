package diff

import (
	"errors"
	"reflect"
	"strings"
)

// Reason constants
type Reason int

const (
	// DiffOfUnknown unknown diff reason
	DiffOfUnknown Reason = iota
	// DiffOfType different of type
	DiffOfType
	// DiffOfSliceLength different of slice length
	DiffOfSliceLength
	// DiffOfMapLength different of map length
	DiffOfMapLength
	// DiffOfValue different of value
	DiffOfValue
	// DiffOfLeftNoValue different cause left has no value
	DiffOfLeftNoValue
	// DiffOfRightNoValue different cause right has no value
	DiffOfRightNoValue
	// DiffOfLeftElemRemoved different cause one element of left is removed
	DiffOfLeftElemRemoved
	// DiffOfRightElemAdded different cause one element of right is added
	DiffOfRightElemAdded
)

func (re Reason) String() string {
	switch re {
	case DiffOfType:
		return "Diff Type"
	case DiffOfSliceLength:
		return "Diff Slice Length"
	case DiffOfMapLength:
		return "Diff Map Length"
	case DiffOfValue:
		return "Diff Value"
	case DiffOfLeftNoValue:
		return "Diff Left No Value"
	case DiffOfRightNoValue:
		return "Diff Right No Value"
	case DiffOfLeftElemRemoved:
		return "Diff Left Elem Removed"
	case DiffOfRightElemAdded:
		return "Diff Right Elem Added"
	}
	return "Diff Unknown"
}

// Callback invoked when left is different with right
type Callback func(*D) (shouldContinue bool)

type callbackD func(path string, reason Reason, leftV reflect.Value, rightV reflect.Value) (shouldContinue bool)

// Differ with compare functions
type Differ struct {
	// the cmpFunc should be func(left,right customType) bool
	cmpPathFuncs map[pathType]reflect.Value
	// the cmpFunc should be func(left,right customType) bool
	cmpFuncs map[reflect.Type]reflect.Value
	// the cmpKindFunc should be func(left,right primitiveKind) bool
	cmpKindFuncs map[reflect.Kind]reflect.Value
	// typeIDFunc should be func(v cumstomType) (id string)
	typeIDFuncs map[reflect.Type]reflect.Value
	kindIDFuncs map[reflect.Kind]reflect.Value
	omitPaths   map[string]bool
	omitPrefix  map[string]bool
}

type pathType struct {
	P string
	T reflect.Type
}
type differ struct {
	*Differ
	// Callback when diff value found
	Callback        callbackD
	differenceExist bool
	typeCache       *typeIDCache
	pathToType      map[string]reflect.Type
}

// New differ with default config
func New() *Differ {
	df := &Differ{
		cmpFuncs:     make(map[reflect.Type]reflect.Value),
		cmpPathFuncs: make(map[pathType]reflect.Value),
		cmpKindFuncs: make(map[reflect.Kind]reflect.Value),
		typeIDFuncs:  make(map[reflect.Type]reflect.Value),
		kindIDFuncs:  make(map[reflect.Kind]reflect.Value),
		omitPaths:    make(map[string]bool),
		omitPrefix:   make(map[string]bool),
	}
	df.registDefaultCmpFuncs()
	return df
}

func newDiffer(d *Differ, fn Callback) *differ {
	_diff := &differ{
		Differ:     d,
		typeCache:  newTypeIDCache(),
		pathToType: make(map[string]reflect.Type),
	}
	wfn := func(path string, reason Reason, leftV reflect.Value, rightV reflect.Value) (shouldContinue bool) {
		if d.isOmit(path) {
			return true
		}
		_diff.differenceExist = true
		_d := buildD(path, reason, leftV, rightV)
		if t, ok := _diff.getPathToType(path); ok && t.Kind() == reflect.Ptr {
			if leftV.Kind() != reflect.Ptr && leftV.Kind() != reflect.Interface && leftV.IsValid() {
				nLeft := reflect.New(t.Elem())
				nLeft.Elem().Set(leftV)
				_d.LeftV = nLeft
			}
			if rightV.Kind() != reflect.Ptr && rightV.Kind() != reflect.Interface && rightV.IsValid() {
				nRight := reflect.New(t.Elem())
				nRight.Elem().Set(rightV)
				_d.RightV = nRight
			}
		}
		return fn(_d)
	}
	_diff.Callback = wfn
	return _diff
}

// OmitPath would be skipped by differ, the path can be absolute path .A.B.C or last path step D or slice fuzzy path .A[*].E
func (df *Differ) OmitPath(list ...string) {
	for _, p := range list {
		if isPathPrefix(p) {
			df.omitPrefix[getPathPrefix(p)] = true
		} else {
			df.omitPaths[p] = true
		}
	}
}

func (df *Differ) isOmit(p string) bool {
	if df.omitPaths[p] {
		return true
	}
	if df.omitPaths[LastNodeOfPath(p)] {
		return true
	}
	p = replaceSliceIndexToStar(p)
	if df.omitPaths[p] {
		return true
	}
	for prefix := range df.omitPrefix {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// RegistCompareFunc the cmpFunc should be func(left,right customType) bool
func (df *Differ) RegistCompareFunc(fn interface{}) error {
	err := errors.New("the cmpFunc should be func(left,right customType) bool")
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != 2 {
		return err
	}
	if f.Type().In(1) != f.Type().In(0) {
		return err
	}
	if f.Type().NumOut() != 1 {
		return err
	}
	if f.Type().Out(0) != reflect.TypeOf(true) {
		return err
	}
	df.cmpFuncs[f.Type().In(1)] = f
	return nil
}

// RegistPathCompareFunc the cmpFunc should be func(left,right customType) bool
func (df *Differ) RegistPathCompareFunc(path string, fn interface{}) error {
	err := errors.New("the cmpFunc should be func(left,right customType) bool")
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != 2 {
		return err
	}
	if f.Type().In(1) != f.Type().In(0) {
		return err
	}
	if f.Type().NumOut() != 1 {
		return err
	}
	if f.Type().Out(0) != reflect.TypeOf(true) {
		return err
	}
	df.cmpPathFuncs[buildPathType(path, f.Type().In(0))] = f
	return nil
}

// RegistCompareKindFunc the cmpKindFunc should be func(left,right primitiveKind) bool
func (df *Differ) RegistCompareKindFunc(fn interface{}) error {
	err := errors.New("the cmpKindFunc should be func(left,right primitiveKind) bool")
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != 2 {
		return err
	}
	if f.Type().NumOut() != 1 {
		return err
	}
	if f.Type().In(0) != f.Type().In(1) {
		return err
	}
	if f.Type().Out(0) != reflect.TypeOf(true) {
		return err
	}
	df.cmpKindFuncs[f.Type().In(0).Kind()] = f
	return nil
}

// RegistIDFunc should be func(v cumstomType) (id string)
func (df *Differ) RegistIDFunc(fn interface{}) error {
	err := errors.New("fn should be func(v cumstomType) (id string)")
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != 1 || f.Type().NumOut() != 1 {
		return err
	}
	if f.Type().Out(0) != reflect.TypeOf("") {
		return err
	}
	df.typeIDFuncs[f.Type().In(0)] = f
	return nil
}

// RegistKindIDFunc should be func(v cumstomType) (id string)
func (df *Differ) RegistKindIDFunc(fn interface{}) error {
	err := errors.New("fn should be func(v cumstomType) (id string)")
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != 1 || f.Type().NumOut() != 1 {
		return err
	}
	if f.Type().Out(0) != reflect.TypeOf("") {
		return err
	}
	df.kindIDFuncs[f.Type().In(0).Kind()] = f
	return nil
}

/* private constants */
const (
	_SPLITTOR = "."
	_ROOT     = _SPLITTOR
	_ZERO     = "_ZERO_VALUE_"
)

/* private methods */

func (df *Differ) registDefaultCmpFuncs() {
	// kind compare
	df.RegistCompareKindFunc(CmpBool)
	df.RegistCompareKindFunc(CmpInt)
	df.RegistCompareKindFunc(CmpInt8)
	df.RegistCompareKindFunc(CmpInt16)
	df.RegistCompareKindFunc(CmpInt32)
	df.RegistCompareKindFunc(CmpInt64)
	df.RegistCompareKindFunc(CmpUint)
	df.RegistCompareKindFunc(CmpUint8)
	df.RegistCompareKindFunc(CmpUint16)
	df.RegistCompareKindFunc(CmpUint32)
	df.RegistCompareKindFunc(CmpUint64)
	df.RegistCompareKindFunc(CmpUintptr)
	df.RegistCompareKindFunc(CmpFloat32)
	df.RegistCompareKindFunc(CmpFloat64)
	df.RegistCompareKindFunc(CmpString)
	df.RegistCompareKindFunc(CmpUnsafePointer)

	df.RegistCompareFunc(CmpTime)
	df.RegistCompareFunc(CmpTimePtr)

	df.RegistKindIDFunc(IDOfBool)
	df.RegistKindIDFunc(IDOfInt)
	df.RegistKindIDFunc(IDOfInt8)
	df.RegistKindIDFunc(IDOfInt16)
	df.RegistKindIDFunc(IDOfInt32)
	df.RegistKindIDFunc(IDOfInt64)
	df.RegistKindIDFunc(IDOfUint)
	df.RegistKindIDFunc(IDOfUint8)
	df.RegistKindIDFunc(IDOfUint16)
	df.RegistKindIDFunc(IDOfUint32)
	df.RegistKindIDFunc(IDOfUint64)
	df.RegistKindIDFunc(IDOfUintptr)
	df.RegistKindIDFunc(IDOfString)
	df.RegistKindIDFunc(IDOfFloat32)
	df.RegistKindIDFunc(IDOfFloat64)
}

func (df *Differ) canCmpType(path string, t reflect.Type) bool {
	_, ok := df.cmpFuncs[t]
	_, ok1 := df.cmpPathFuncs[buildPathType(path, t)]
	return ok || ok1
}

func (df *Differ) getCmpTypeFn(path string, t reflect.Type) reflect.Value {
	if fn, ok := df.cmpPathFuncs[buildPathType(path, t)]; ok {
		return fn
	}
	return df.cmpFuncs[t]
}

func (df *differ) setPathToType(path string, tp reflect.Type) {
	path = replaceSliceIndexToStar(path)
	if _, ok := df.pathToType[path]; ok {
		return
	}
	df.pathToType[path] = tp
}

func (df *differ) forceSetPathToType(path string, tp reflect.Type) {
	path = replaceSliceIndexToStar(path)
	df.pathToType[path] = tp
}

func (df *differ) getPathToType(path string) (reflect.Type, bool) {
	path = replaceSliceIndexToStar(path)
	t, ok := df.pathToType[path]
	return t, ok
}

func (df *differ) cmpByType(steps []string, t reflect.Type, lv, rv reflect.Value) bool {
	fn := df.getCmpTypeFn(buildPath(steps), t)
	if !lv.IsValid() || !rv.IsValid() {
		if !lv.IsValid() && rv.IsValid() {
			return df.Callback(buildPath(steps), DiffOfLeftNoValue, lv, rv)
		} else if lv.IsValid() && !rv.IsValid() {
			return df.Callback(buildPath(steps), DiffOfRightNoValue, lv, rv)
		}
		return true
	}
	out := fn.Call([]reflect.Value{lv, rv})
	equal := out[0].Bool()
	if !equal {
		return df.Callback(buildPath(steps), DiffOfValue, lv, rv)
	}
	return true
}

func (df *differ) cmpByKind(steps []string, kind reflect.Kind, lk, rk reflect.Value) bool {
	fn := df.cmpKindFuncs[kind]
	if !lk.IsValid() || !rk.IsValid() {
		if !lk.IsValid() && rk.IsValid() {
			return df.Callback(buildPath(steps), DiffOfLeftNoValue, lk, rk)
		} else if lk.IsValid() && !rk.IsValid() {
			return df.Callback(buildPath(steps), DiffOfRightNoValue, lk, rk)
		}
		return true
	}
	out := fn.Call([]reflect.Value{lk.Convert(fn.Type().In(0)), rk.Convert(fn.Type().In(0))})
	equal := out[0].Bool()
	if !equal {
		return df.Callback(buildPath(steps), DiffOfValue, lk, rk)
	}
	return true
}

func cmpVal(df *differ, steps []string, t reflect.Type, lv, rv reflect.Value) bool {
	df.setPathToType(buildPath(steps), t)

	if df.canCmpType(buildPath(steps), t) {
		return df.cmpByType(steps, t, lv, rv)
	}
	switch t.Kind() {
	case reflect.String:
		return df.cmpByKind(steps, t.Kind(), lv, rv)
	case reflect.Bool:
		return df.cmpByKind(steps, t.Kind(), lv, rv)
	case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int32, reflect.Int64:
		return df.cmpByKind(steps, t.Kind(), lv, rv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return df.cmpByKind(steps, t.Kind(), lv, rv)
	case reflect.Float32, reflect.Float64:
		return df.cmpByKind(steps, t.Kind(), lv, rv)
	case reflect.Struct:
		return cmpStruct(df, steps, t, lv, rv)
	case reflect.Ptr:
		if !lv.IsNil() && !lv.IsNil() {
			return cmpVal(df, steps, t.Elem(), lv.Elem(), rv.Elem())
		} else if lv.IsNil() && !rv.IsNil() {
			return df.Callback(buildPath(steps), DiffOfLeftNoValue, lv, rv)
		} else if !lv.IsNil() && rv.IsNil() {
			return df.Callback(buildPath(steps), DiffOfRightNoValue, lv, rv)
		}
	case reflect.Map:
		if lv.Type() != rv.Type() {
			return df.Callback(buildPath(steps), DiffOfType, lv, rv)
		}
		if !lv.IsNil() && !lv.IsNil() {
			return cmpMap(df, steps, t.Key(), t.Elem(), lv, rv)
		} else if lv.IsNil() && !rv.IsNil() {
			return df.Callback(buildPath(steps), DiffOfLeftNoValue, lv, rv)
		} else if !lv.IsNil() && rv.IsNil() {
			return df.Callback(buildPath(steps), DiffOfRightNoValue, lv, rv)
		}
	case reflect.Slice, reflect.Array:
		return cmpSlice(df, steps, t.Elem(), lv, rv)
	case reflect.Chan, reflect.Func:
		return true
	case reflect.UnsafePointer:
		return df.cmpByKind(steps, t.Kind(), lv, rv)
	case reflect.Interface:
		if lv.IsNil() == rv.IsNil() {
			if lv.IsNil() {
				return true
			}
			df.forceSetPathToType(buildPath(steps), lv.Elem().Type())
			if lv.Elem().Kind() == reflect.Ptr {
				return cmpVal(df, steps, lv.Elem().Elem().Type(), lv.Elem().Elem(), rv.Elem().Elem())
			} else {
				return cmpVal(df, steps, lv.Elem().Type(), lv.Elem(), rv.Elem())
			}
		} else {
			if lv.IsNil() && !rv.IsNil() {
				df.forceSetPathToType(buildPath(steps), lv.Elem().Type())
				return df.Callback(buildPath(steps), DiffOfLeftNoValue, lv, rv)
			} else if !lv.IsNil() && rv.IsNil() {
				df.forceSetPathToType(buildPath(steps), lv.Elem().Type())
				return df.Callback(buildPath(steps), DiffOfRightNoValue, lv, rv)
			}
		}
	}
	return true
}
