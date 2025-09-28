# Go Clipper2

[![test](https://github.com/bolom009/go-clipper2/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/bolom009/go-clipper2/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/bolom009/go-clipper2/branch/master/graph/badge.svg)](https://codecov.io/gh/bolom009/go-clipper2)

A high-performance pure Go port of
[Clipper2](https://github.com/AngusJohnson/Clipper2), the industry-standard
polygon clipping and offsetting library. Go Clipper2 provides robust geometric
operations with 64-bit integer precision, eliminating floating-point numerical
errors common in computational geometry.

## 🌟 Features

- **🚀 Pure Go Implementation**: Zero C/C++ dependencies for production use
- **🎯 Complete API**: All Clipper2 operations including boolean ops, offsetting, and clipping
- **🧪 Comprehensive Testing**: Property-based testing with different cases
- **📦 Easy Integration**: Simple Go module with clean, idiomatic API

### Prerequisites

- Go 1.23 or later

### Installation

```
go get github.com/bolom009/go-clipper2
```

## 📖 Usage Examples

### Basic Boolean Operations

```go
package main

import (
    "fmt"
    goclipper2 "github.com/bolom009/go-clipper2"
)

func main() {
    // Define two overlapping rectangles
	var (
		subject = goclipper2.Paths64{
			{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		}
		clip = goclipper2.Paths64{
			{{50, 50}, {150, 50}, {150, 150}, {50, 150}},
		}
	)

	// wrapped method booleanOp (simplify usage of lib)
	unionResults := goclipper2.UnionWithClipPaths64(subject, clip, goclipper2.NonZero)
	fmt.Printf("Union area: %v\n", unionResults)
	
	// booleanOp
	unionResults = goclipper2.BooleanOpPaths64(goclipper2.Union, subject, clip, goclipper2.NonZero)
    fmt.Printf("Union area: %v\n", unionResults)
	
	/*
	    goclipper2.UnionPaths64(...)
	    goclipper2.UnionWithClipPaths64(...)
	    goclipper2.IntersectWithClipPaths64(...)
	    goclipper2.DifferenceWithClipPaths64(...)
	    goclipper2.XorWithClipPaths64(...)
	    goclipper2.UnionWithClipPathsD(...)
	    ...
	*/
}
```

## 📊 Implementation Status

| Feature                    | Status |
|----------------------------|--------|
| Boolean Operations {64,D}  | ✅      |
| Union64, UnionD            | ✅      |
| Intersect64, IntersectD    | ✅      |
| Difference64, DifferenceD  | ✅      |
| Xor64, XorD                | ✅      |
| Polygon Offsetting         | ✅      |
| Rectangle Clipping         | ✅      |
| Area Calculation           | ✅      |
| Orientation Detection      | ✅      |
| Path Reversal              | ✅      |
| Minkowski Operations       | ✅      |
| PolyTree, PolyPath {64, D} | ❌      |

**Legend**: ✅ Implemented, ❌ Not implemented, 🚧 In progress

### Performance Tips

- Use integer coordinates when possible (more robust than float64)
- For simple rectangular clipping, use `RectClip64` instead of boolean
  operations
- Pre-simplify complex polygons before operations
- Consider polygon orientation for optimal performance

## 📄 License

This project is licensed under the **Boost Software License 1.0**, the same as
the original Clipper2 library. See [LICENSE](LICENSE) for details.