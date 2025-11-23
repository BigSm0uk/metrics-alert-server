package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	// Запускаем тесты на тестовых данных
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a", "b")
}
