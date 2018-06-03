package main

import (
	"fmt"

	"github.com/slice-d/genzai"
)

type greeting string

func (g greeting) Greet(context moved.Context) {
	moved.SetContextName("Changed")
	fmt.Println(context.Exec("Hi"))
	fmt.Println("Hello Universe")
}

// exported
var Greeter greeting
