package script

import (
	"testing"

	"github.com/yuin/gopher-lua"
)

func Benchmark(t *testing.B) {
	//L := lua.NewState()
	//defer L.Close()
	//if err := L.DoString(`print("hello")`); err != nil {
	//	panic(err)
	//}

	L := lua.NewState()
	defer L.Close()

	fn, err := L.LoadFile("script.lua")
	if err != nil {
		panic(err)
	}


	t.ResetTimer()
	for i:=0; i < t.N; i++ {
		L.Push(fn)
		L.Call(0, lua.MultRet)
		//L.Push(fn)
		//L.PCall(0, lua.MultRet, nil)
		L.Pop(1)
	}
	t.StopTimer()

	//if err := L.DoFile("script.lua"); err != nil {
	//	panic(err)
	//}
}


func Test(t *testing.T) {
	//L := lua.NewState()
	//defer L.Close()
	//if err := L.DoString(`print("hello")`); err != nil {
	//	panic(err)
	//}

	L := lua.NewState()
	defer L.Close()

	fn, err := L.LoadFile("script.lua")
	if err != nil {
		panic(err)
	}


	L.Push(fn)
	L.Call(0, lua.MultRet)
	L.Push(fn)
	L.Call(0, lua.MultRet)
	//L.Push(fn)
	//L.PCall(0, lua.MultRet, nil)


	//if err := L.DoFile("script.lua"); err != nil {
	//	panic(err)
	//}
}