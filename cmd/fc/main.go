package main

import (
	"osscert"

	"github.com/aliyun/fc-runtime-go-sdk/fc"
)

func main() {
	fc.Start(handler)
}

func handler() (string, error) {
	return osscert.Run()
}
