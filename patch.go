package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// D is a single path row
type D struct {
	Path   string
	Reason Reason
	LeftV  reflect.Value
	RightV reflect.Value
}

// Indirect of D
func (d D) Indirect() *D {
	return &D{
		Path:   d.Path,
		Reason: d.Reason,
		LeftV:  reflect.Indirect(d.LeftV),
		RightV: reflect.Indirect(d.RightV),
	}
}

// Patch is result of diff
type Patch struct {
	List []*D
}

// DInterface is interface of D
type DInterface struct {
	Path          string
	Reason        Reason
	LeftV, RightV interface{}
}

// PatchValue is result of diff
type PatchValue struct {
	List []DInterface
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

// Readable string format
func (p *Patch) Readable() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("Patch size: %v\n", p.Size()))
	for i, d := range p.List {
		var datal, datar []byte
		if d.LeftV.IsValid() {
			datal, _ = json.Marshal(d.LeftV.Interface())
		}
		if d.RightV.IsValid() {
			datar, _ = json.Marshal(d.RightV.Interface())
		}
		b.WriteString(fmt.Sprintf("%02d. %s (%s) left=(%v) right=(%v)\n", i+1, d.Path, d.Reason, string(datal), string(datar)))
	}
	b.WriteRune('\n')
	return b.String()
}

// Interface of patch details
func (p *Patch) Interface() PatchValue {
	var list []DInterface
	for _, v := range p.List {
		list = append(list, DInterface{
			Path:   v.Path,
			Reason: v.Reason,
			LeftV:  v.LeftV.Interface(),
			RightV: v.RightV.Interface(),
		})
	}
	return PatchValue{List: list}
}
