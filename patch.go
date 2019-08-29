package diff

import "reflect"

// D is a single path row
type D struct {
	Path   string
	Reason Reason
	LeftV  reflect.Value
	RightV reflect.Value
}

// Patch is result of diff
type Patch struct {
	List []*D
}

// Size of patch
func (p *Patch) Size() int {
	return len(p.List)
}

// IsEmpty patch
func (p *Patch) IsEmpty() bool {
	return p.Size() == 0
}

func (p *Patch) add(d *D) *Patch {
	p.List = append(p.List, d)
	return p
}
