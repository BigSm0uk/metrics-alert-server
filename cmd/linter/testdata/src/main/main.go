package main

import (
	"log"
	"os"
)

// Этот файл проверяет, что в main пакете main не должно быть предупреждений
// для os.Exit и log.Fatal внутри функции main

func main() {
	// В функции main пакета main os.Exit и log.Fatal разрешены
	if false {
		os.Exit(0) // это OK
	}
	if false {
		log.Fatal("error") // это OK
	}

	// Но panic всё равно должен быть обнаружен
	if false {
		panic("error") // want "использование встроенной функции panic"
	}
}

func helper() {
	// Вне main всё равно не разрешено
	os.Exit(1)         // want "вызов os.Exit вне функции main пакета main"
	log.Fatal("error") // want "вызов log.Fatal вне функции main пакета main"
	panic("error")     // want "использование встроенной функции panic"
}
