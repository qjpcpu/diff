package diff

import (
	"bytes"
	"encoding/gob"
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
		b.WriteString(fmt.Sprintf("%02d. %s (%s) left=(%v) right=(%v)\n", i+1, d.Path, d.Reason, string(datal), string(datar)))
	}
	b.WriteRune('\n')
	return b.String()
}

type transD struct {
	Path          string
	Reason        Reason
	LeftV, RightV interface{}
}

type transPatch struct {
	List []transD
}

func (d D) toTrans() transD {
	return transD{
		Path:   d.Path,
		Reason: d.Reason,
		LeftV:  d.LeftV.Interface(),
		RightV: d.RightV.Interface(),
	}
}

func (d transD) toD() *D {
	return &D{
		Path:   d.Path,
		Reason: d.Reason,
		LeftV:  reflect.ValueOf(d.LeftV),
		RightV: reflect.ValueOf(d.RightV),
	}
}
func (p *Patch) toTrans() (tp transPatch) {
	for _, d := range p.List {
		tp.List = append(tp.List, d.toTrans())
	}
	return
}
func (p transPatch) toPatch() *Patch {
	pa := &Patch{}
	for _, d := range p.List {
		pa.List = append(pa.List, d.toD())
	}
	return pa
}

// Encode patch to bytes
func (p *Patch) Encode() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(p.toTrans())
	return b.Bytes(), err
}

// Decode patch from bytes
func (p *Patch) Decode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	var tp transPatch
	if err := dec.Decode(&tp); err != nil {
		return err
	}
	pch := tp.toPatch()
	*p = *pch
	return nil
}
