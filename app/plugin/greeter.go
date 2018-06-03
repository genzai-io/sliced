package main

import (
	"fmt"
	"os"
	"plugin"
	"time"

	"github.com/slice-d/genzai"
)

type Greeter interface {
	Greet(context moved.Context)
}

func main() {
	fmt.Println(os.Args)
	// determine module to load
	//lang := "english"
	//if len(os.Args) == 2 {
	//	lang = os.Args[1]
	//}
	//var mod string
	//switch lang {
	//case "english":
	//	mod = "./eng/eng.so"
	//case "chinese":
	//	mod = "./chi/chi.so"
	//default:
	//	fmt.Println("don't speak that language")
	//	os.Exit(1)
	//}

	for {
		Load("./eng/eng.so")
		Load("./chi/chi.so")
		time.Sleep(time.Millisecond * 1000)
	}
}

func Load(mod string) {
	// load module
	// 1. open the so file to load the symbols
	plug, err := plugin.Open(mod)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 2. look up a symbol (an exported function or variable)
	// in this case, variable Greeter
	symGreeter, err := plug.Lookup("Greeter")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 3. Assert that loaded symbol is of a desired type
	// in this case interface type Greeter (defined above)
	var greeter Greeter
	greeter, ok := symGreeter.(Greeter)
	if !ok {
		fmt.Println("unexpected type from module symbol")
		os.Exit(1)
	}

	moved.SetContextName("Main App")

	// 4. use the module
	greeter.Greet(moved.PluginCtx)

	fmt.Println(moved.GetContextName())
}