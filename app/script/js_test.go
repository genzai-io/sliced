package script

import (
	"fmt"
	"testing"
	"time"

	"github.com/robertkrimen/otto"
)

func TestJS(t *testing.T) {
	vm := otto.New()

	vm.Set("sayHello", func(call otto.FunctionCall) otto.Value {
		fmt.Printf("Hello, %s.\n", call.Argument(0).String())
		return otto.Value{}
	})

	vm.Set("now", func(call otto.FunctionCall) otto.Value {
		//fmt.Printf("Hello, %s.\n", call.Argument(0).String())
		val, _ := otto.ToValue(time.Now().String())
		return val
	})

	//vm.Run(`
	//abc = 2 + 2;
	//sayHello('Hi');
	//console.log("The value of abc is " + abc); // 4`)

	script, err := vm.Compile("", `
    abc = 2 + 2;
	sayHello('Hi');
console.log(now());
console.log(Object.getOwnPropertyNames(Object.prototype));
    console.log("The value of abc is " + abc); // 4`)
	if err != nil {
		t.Fatal(err)
	}

	vm.Run(script)
}

func BenchmarkJS(t *testing.B) {
	vm := otto.New()

	vm.Set("sayHello", func(call otto.FunctionCall) otto.Value {
		//fmt.Printf("Hello, %s.\n", call.Argument(0).String())
		return otto.Value{}
	})

	// streams.get('mystream')
	// stream.cursor()
	// cursor.go(
	script, err := vm.Compile("", "var x = 0; if (x == 0) {} else {}")
	if err != nil {
		t.Fatal(err)
	}
	_ = script

	t.ResetTimer()
	t.StartTimer()
	for i := 0; i < t.N; i++ {
		vm.Run(script)
	}
	t.StopTimer()
}
