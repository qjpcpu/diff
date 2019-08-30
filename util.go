package diff

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// IsPrimitiveType is simple types, bool,intx,uintx,floatx
func IsPrimitiveType(tp reflect.Type) bool {
	kind := tp.Kind()
	return (reflect.Bool <= kind && kind <= reflect.Float64) || kind == reflect.String
}

func buildPath(steps []string) string {
	var list []string
	for i, s := range steps {
		if isIndexToken(s) && i > 0 {
			list = append(list, s)
		} else {
			list = append(list, _SPLITTOR, s)
		}
	}
	return strings.Join(list, "")
}

// is like [number]
func isIndexToken(s string) bool {
	token := []byte(s)
	if len(token) < 3 || token[0] != '[' || token[len(token)-1] != ']' {
		return false
	}
	// filt [00????]
	if token[1] == '0' && token[2] == '0' {
		return false
	}
	for i := 1; i < len(token)-1; i++ {
		if token[i] < '0' || token[i] > '9' {
			return false
		}
	}
	return true
}

func replaceSliceIndexToStar(p string) string {
	token := []byte(p)
	var bracket bool
	var moreThanOne bool
	for i := 0; i < len(token); i++ {
		if token[i] == '[' {
			bracket = true
			moreThanOne = true
			continue
		}
		if bracket && token[i] >= '0' && token[i] <= '9' {
			token[i] = '*'
		}
		if token[i] == ']' {
			bracket = false
		}
	}
	if !moreThanOne {
		return p
	}
	var offset int
	for i, b := range token {
		if b == '*' {
			if i == 0 || token[i-1] != '*' {
				token[i-offset] = b
			} else {
				offset++
			}
		} else {
			token[i-offset] = b
		}
	}
	return string(token[:len(token)-offset])
}

func buildIndexStep(i int) string {
	return "[" + strconv.FormatInt(int64(i), 10) + "]"
}

func appendPath(steps []string, tokens ...string) []string {
	return append(steps, tokens...)
}

func isExported(fieldName string) bool {
	return len(fieldName) > 0 && (fieldName[0] >= 'A' && fieldName[0] <= 'Z')
}

func buildPathType(path string, t reflect.Type) pathType {
	return pathType{P: path, T: t}
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func defaultValue(t reflect.Type) reflect.Value {
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem())
	}
	return reflect.New(t).Elem()
}

func valueToString(fromv reflect.Value) string {
	switch fromv.Kind() {
	case reflect.String:
		return fromv.String()
	case reflect.Bool:
		return strconv.FormatBool(fromv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(fromv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(fromv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(fromv.Float(), 'f', 6, 64)
	}
	return fmt.Sprint(fromv.Interface())
}

func buildD(path string, reason Reason, leftV reflect.Value, rightV reflect.Value) *D {
	return &D{
		Path:   path,
		Reason: reason,
		LeftV:  leftV,
		RightV: rightV,
	}
}
