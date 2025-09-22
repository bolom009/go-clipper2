package go_clipper2

type Clipper64 struct {
	ClipperBase
}

func NewClipper64() *Clipper64 {
	return &Clipper64{
		ClipperBase{},
	}
}
