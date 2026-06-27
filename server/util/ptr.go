package util

// Safe dereferencing of pointers to make
// our core logic more readable.
// If the ptr is nil, def (a default value) is returned.
func Deref[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}
