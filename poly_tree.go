package go_clipper2

import "fmt"

type PolyTreeD PolyPathD

func NewPolyTreeD() *PolyTreeD {
	return &PolyTreeD{
		PolyPathBase: NewPolyPathBase(nil),
	}
}

type PolyPathD struct {
	*PolyPathBase
}

type PolyTree64 PolyPath64

func NewPolyTree64() *PolyTree64 {
	return &PolyTree64{
		PolyPathBase: NewPolyPathBase(nil),
	}
}

type PolyPath64 struct {
	*PolyPathBase
}

// func (p *PolyBase) AddChild(path64 Path64) *PolyBase {}

type PolyPathBase struct {
	parent *PolyPathBase
	childs []*PolyPathBase

	scale float64
}

func NewPolyPathBase(parent *PolyPathBase) *PolyPathBase {
	return &PolyPathBase{
		parent,
		make([]*PolyPathBase, 0),
		0,
	}
}

func (p *PolyPathBase) AddChild(pth Path64) *PolyPathBase {
	child := NewPolyPathBase(p)
	p.childs = append(p.childs, child)
	return child
}

func (p *PolyPathBase) Count() int {
	return len(p.childs)
}

func (p *PolyPathBase) SetScale(scale float64) {
	p.scale = scale
}

func (p *PolyPathBase) Scale() float64 {
	return p.scale
}

func (p *PolyPathBase) GetChildren() []*PolyPathBase {
	return p.childs
}

func (p *PolyPathBase) IsHole() bool {
	lvl := p.Level()
	return lvl != 0 && (lvl&1) == 0
}

func (p *PolyPathBase) Level() int {
	result := 0
	pp := p.parent
	for pp != nil {
		result++
		pp = pp.parent
	}

	return result
}

func (p *PolyPathBase) Clear() {
	p.childs = p.childs[:0]
}

func (p *PolyPathBase) ToString() string {
	if p.Level() > 0 {
		return ""
	}

	plural := "s"
	if len(p.childs) == 1 {
		plural = ""
	}

	result := fmt.Sprintf("Polytree with %d polygon%v.\n", len(p.childs), plural)
	for i := 0; i < len(p.childs); i++ {
		if len(p.childs[i].childs) > 0 {
			result += p.childs[i].ToStringInternal(i, 1)
		}
	}

	return result + "\n"
}

func (p *PolyPathBase) ToStringInternal(idx, level int) string {
	var (
		result = ""
		plural = "s"
	)

	if len(p.childs) == 1 {
		plural = ""
	}

	if (level & 1) == 0 {
		result += fmt.Sprintf("%*v +- hole (%v) contains %v nested polygon%v.\n", level*2, "", idx, len(p.childs), plural)
	} else {
		result += fmt.Sprintf("%*v +- polygon (%v) contains %v hole%v.\n", level*2, "", idx, len(p.childs), plural)
	}

	for i := 0; i < len(p.childs); i++ {
		if len(p.childs[i].childs) > 0 {
			result += p.childs[i].ToStringInternal(i, level+1)
		}
	}

	return result
}
