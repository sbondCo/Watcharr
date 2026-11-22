// This package is for representing 3 states in a more readable/self-documenting
// manner.
// It is the alternative to using `*bool`, which helps us avoid using pointers
// (which avoids runtime panics if we forget to nil check and possibly other
// logical errors).
// Obviously if we are unmarshalling JSON, you gotta still use *bool if you
// need to know the differece between set/unset. Use this package for improving
// internal communications only.
package tri

type State int

// NOTE: The `unset` const **MUST** be `0`, because that is the default value
// for an int in Go.
// For ex., if `Unset` was `-1`, then a non-initialized var/property, would
// default to `0` (aka False in this scenario). That's why it starts at `0`.

const (
	Unset State = iota
	False
	True
)

// Turn a `State` into a `bool`.
//
// `true` will only ever be returned if `s` is definitely True!
func ToBool(s State) bool {
	if s == True {
		return true
	}
	return false
}

// Turn a `bool` into a `State`.
func FromBool(s bool) State {
	if s {
		return True
	}
	return False
}
