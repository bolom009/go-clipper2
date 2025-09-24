package go_clipper2

import "fmt"

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

func binarySearch(arr []int64, target int64) int {
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
	return -1
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
