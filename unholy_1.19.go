//go:build !go1.20

package shift

import "unsafe"

// bytesToString converts provided bytes to string without incurring additional allocations using unsafe type casting.
func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
