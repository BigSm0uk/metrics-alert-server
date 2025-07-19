package util

type Number interface {
	int64 |
		float64 | int | float32
}

func GetDefault[T Number](value *T) T {
	if value == nil {
		return 0
	}
	return *value
}
