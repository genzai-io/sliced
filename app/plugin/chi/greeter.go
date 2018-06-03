package main

import (
	"fmt"

	"github.com/slice-d/genzai"
)

type greeting string

func (g greeting) Greet(context moved.Context) {
	fmt.Println(context.Exec("Hi"))
	fmt.Println("你好宇宙")
}

var Greeter greeting
