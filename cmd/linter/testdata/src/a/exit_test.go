package a

import (
	"log"
	"os"
)

// Тесты для проверки обнаружения os.Exit и log.Fatal вне main

func badExit() {
	os.Exit(1) // want "вызов os.Exit вне функции main пакета main"
}

func badFatal() {
	log.Fatal("error") // want "вызов log.Fatal вне функции main пакета main"
}

func badFatalf() {
	log.Fatalf("error: %s", "test") // want "вызов log.Fatalf вне функции main пакета main"
}

func badFatalln() {
	log.Fatalln("error") // want "вызов log.Fatalln вне функции main пакета main"
}

type MyStruct struct{}

func (m *MyStruct) method() {
	os.Exit(1) // want "вызов os.Exit вне функции main пакета main"
}

func init() {
	log.Fatal("init") // want "вызов log.Fatal вне функции main пакета main"
}
