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

// Readable string format
func (p *Patch) Readable() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("Patch size: %v\n", p.Size()))
	for i, d := range p.List {
		datal, _ := json.Marshal(d.LeftV.Interface())
		datar, _ := json.Marshal(d.RightV.Interface())
		b.WriteString(fmt.Sprintf("%2d. %s (%s) left=(%v) right=(%v)\n", i+1, d.Path, d.Reason, string(datal), string(datar)))
	}
	b.WriteRune('\n')
	return b.String()
}
