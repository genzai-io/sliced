package script

import (
	"fmt"
	"testing"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/ry/v8worker2"
	"sync"
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

var worker *v8worker2.Worker
var once *sync.Once = &sync.Once{}
func BenchmarkV8(t *testing.B) {
	//worker := v8worker2.New(func(msg []byte) []byte {
	//	//if len(msg) != 5 {
	//	//	t.Fatal("bad msg", msg)
	//	//}
	//	//recvCount++
	//	return nil
	//})

	once.Do(func() {
		worker = v8worker2.New(func(msg []byte) []byte {
			//if len(msg) != 5 {
			//	t.Fatal("bad msg", msg)
			//}
			//recvCount++
			return nil
		})
		err := worker.Load("send.js", `V8Worker2.send(new ArrayBuffer(5));`)
		err = worker.Load("receive.js", `V8Worker2.recv(function(msg) {});`)
		_ = err
		if err != nil {
			t.Fatal(err)
		}
	})


	//err := worker.Load("codeWithRecv.js", `
	//	V8Worker2.recv(function(msg) {
	//	});
	//`)
//	err := worker.Load("codeWithRecv.js", `
//		V8Worker2.recv(function(msg) {
////V8Worker2.send(new ArrayBuffer(5));
////V8Worker2.print(msg);
//			//V8Worker2.print("TestBasic recv byteLength", msg.byteLength);
//		});
//	`)


	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		err := worker.SendBytes([]byte("hii"))
		if err != nil {
			t.Fatal(err)
		}
	}
	t.StopTimer()
		//worker.TerminateExecution()
		//worker.Dispose()

}

func TestV8(t *testing.T) {
	recvCount := 0
	worker := v8worker2.New(func(msg []byte) []byte {
		if len(msg) != 5 {
			t.Fatal("bad msg", msg)
		}
		recvCount++
		return nil
	})

	err := worker.Load("codeWithRecv.js", `
		V8Worker2.recv(function(msg) {
			V8Worker2.print("TestBasic recv byteLength", msg.byteLength);
			if (msg.byteLength !== 3) {
				throw Error("bad message");
			}
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	err = worker.SendBytes([]byte("hii"))
	if err != nil {
		t.Fatal(err)
	}
	codeWithSend := `
		V8Worker2.send(new ArrayBuffer(5));
		V8Worker2.send(new ArrayBuffer(5));
	`
	err = worker.Load("codeWithSend.js", codeWithSend)
	if err != nil {
		t.Fatal(err)
	}

	if recvCount != 2 {
		t.Fatal("bad recvCount", recvCount)
	}
}

func TestV82(t *testing.T) {
	recvCount := 0
	worker := v8worker2.New(func(msg []byte) []byte {
		//if len(msg) != 5 {
		//	t.Fatal("bad msg", msg)
		//}
		recvCount++
		return nil
	})

	err := worker.Load("codeWithRecv.js", `
		V8Worker2.recv(function(msg) {
V8Worker2.send(new ArrayBuffer(5));
//V8Worker2.print(msg);
			//V8Worker2.print("TestBasic recv byteLength", msg.byteLength);
		});
	`)
	if err != nil {
		t.Fatal(err)
	}
	err = worker.SendBytes([]byte("hii"))
	if err != nil {
		t.Fatal(err)
	}
	//codeWithSend := `
	//	V8Worker2.send(new ArrayBuffer(5));
	//	V8Worker2.send(new ArrayBuffer(5));
	//`
	//err = worker.Load("codeWithSend.js", codeWithSend)
	//if err != nil {
	//	t.Fatal(err)
	//}

	if recvCount != 2 {
		//t.Fatal("bad recvCount", recvCount)
	}
}