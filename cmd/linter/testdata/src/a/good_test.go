package a

import (
	"fmt"
)

// Положительные примеры - код, который не должен вызывать предупреждений

func goodFunction() {
	fmt.Println("This is OK")
}

func goodError() error {
	return fmt.Errorf("error")
}

func goodRecover() {
	// recover вызывать можно
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered")
		}
	}()
}
