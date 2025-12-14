package a

// Тесты для проверки обнаружения использования panic

func badPanic() {
	panic("error") // want "использование встроенной функции panic"
}

func anotherBadPanic() {
	if true {
		panic("test") // want "использование встроенной функции panic"
	}
}

func nestedPanic() {
	func() {
		panic("nested") // want "использование встроенной функции panic"
	}()
}
