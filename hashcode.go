package go_clipper2

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
)

const (
	prime2 uint32 = 2246822519
	prime3 uint32 = 3266489917
	prime4 uint32 = 668265263
	prime5 uint32 = 374761393
)

var (
	globalSeed uint32
	seedOnce   sync.Once
)

func initGlobalSeed() {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		// fallback deterministic seed on error
		globalSeed = 0xDEADBEEF
		return
	}
	globalSeed = binary.LittleEndian.Uint32(b[:])
}

// CombineHashes combine two values' hashcodes (similar to C# generic Combine<T1,T2>)
func CombineHashes(hc1 uint32, hc2 uint32) int32 {
	seedOnce.Do(initGlobalSeed)
	hash := mixEmptyState()
	// add length in bytes: two uint32 = 8
	hash += 8
	hash = queueRound(hash, hc1)
	hash = queueRound(hash, hc2)
	hash = mixFinal(hash)
	return int32(hash)
}

// CombineValues helper for when you have interface{} values and want to combine like C# Combine<T1,T2>
func CombineValues(v1 interface{}, v2 interface{}) int32 {
	var hc1 uint32 = 0
	var hc2 uint32 = 0
	if v1 != nil {
		hc1 = uint32(hashOf(v1))
	}
	if v2 != nil {
		hc2 = uint32(hashOf(v2))
	}
	return CombineHashes(hc1, hc2)
}

// queueRound implements RotateLeft(hash + (queuedValue * Prime3), 17) * Prime4
func queueRound(hash uint32, queuedValue uint32) uint32 {
	return rotateLeft(hash+(queuedValue*prime3), 17) * prime4
}

func mixEmptyState() uint32 {
	seedOnce.Do(initGlobalSeed)
	return globalSeed + prime5
}

func mixFinal(hash uint32) uint32 {
	hash ^= hash >> 15
	hash *= prime2
	hash ^= hash >> 13
	hash *= prime3
	hash ^= hash >> 16
	return hash
}

func rotateLeft(value uint32, offset uint) uint32 {
	return (value << offset) | (value >> (32 - offset))
}

// hashOf provides a simple hashcode for some common types to mimic C# GetHashCode behaviour.
// You can extend this as needed for your types.
func hashOf(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t ^ (t >> 32))
	case uint32:
		return int(t)
	case uint64:
		return int(t ^ (t >> 32))
	case string:
		// simple FNV-1a 32-bit
		var h uint32 = 2166136261
		for i := 0; i < len(t); i++ {
			h ^= uint32(t[i])
			h *= 16777619
		}
		return int(h)
	default:
		// fallback: use address (not stable) or zero
		return 0
	}
}
