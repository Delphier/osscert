package main

import (
	"fmt"
	"log"
	"osscert"
)

func main() {
	rst, err := osscert.Run()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(rst)
}
