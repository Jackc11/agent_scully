package utils

// ptr returns a pointer to any literal value (Go 1.18+)
func ToPtr[T any](v T) *T {
	return &v
}
