package util

// DerefString returns the dereferenced value of s, or defaultVal if s is nil.
func DerefString(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

// DerefInt returns the dereferenced value of i, or defaultVal if i is nil.
func DerefInt(i *int, defaultVal int) int {
	if i == nil {
		return defaultVal
	}
	return *i
}

// DerefInt64 returns the dereferenced value of i, or defaultVal if i is nil.
func DerefInt64(i *int64, defaultVal int64) int64 {
	if i == nil {
		return defaultVal
	}
	return *i
}

// DerefBool returns the dereferenced value of b, or defaultVal if b is nil.
func DerefBool(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// Ptr returns a pointer to the value v.
func Ptr[T any](v T) *T {
	return &v
}
