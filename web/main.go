package main

import (
	"fmt"
	"goProject/web/http"
)

func main() {
	engine := http.NewEngine()
	err := engine.Run("9999")
	if err != nil {
		fmt.Printf("run err: %v\n", err)
		return
	}
}
