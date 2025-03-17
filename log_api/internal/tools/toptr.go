package tools

func ToPtr[T comparable](v T) *T {
	return &v
}
