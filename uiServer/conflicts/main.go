package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, Monaco!")
	if len(os.Args) > 1 {
		fmt.Println("Argumento:", os.Args[1])
	}
}