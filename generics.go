package go_clipper2

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

func absInt[T constraints.Signed | constraints.Float](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

func sqr[T constraints.Signed | constraints.Float](val T) T {
	return val * val
}

func ReversePath[T any](p []T) []T {
	n := len(p)
	if n == 0 {
		return []T{}
	}

	rp := make([]T, n)
	for i := 0; i < n; i++ {
		rp[i] = p[n-1-i]
	}
	return rp
}

func binarySearch[T constraints.Ordered](arr []T, target T) int {
	low := 0
	high := len(arr) - 1

	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -(low + 1)
}

func insertAtIndex[T any](slice []T, index int, value T) ([]T, error) {
	if index < 0 || index > len(slice) {
		return slice, fmt.Errorf("index out of bounds")
	}

	slice = append(slice, *new(T))
	copy(slice[index+1:], slice[index:])

	slice[index] = value

	return slice, nil
}
