package diff

import "time"

func CmpTime(path string, left, right time.Time) bool {
	return left.UnixNano() == right.UnixNano()
}

func CmpTimePtr(path string, left, right *time.Time) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return CmpTime(path, *left, *right)
}
