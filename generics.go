package go_clipper2

type Numeric interface {
	float64 | int64
}

type IntNumeric interface {
	int | int64 | int32
}

func absInt[T IntNumeric](a T) T {
	if a < 0 {
		return -a
	}
	return a
}
