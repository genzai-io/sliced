package main

import "github.com/genzai-io/sliced/common/redbench"

func main() {
	redbench.Bench("SET", "127.0.0.1:6379", nil, nil, func(buf []byte) []byte {
		return redbench.AppendCommand(buf, "SET", "key:string", "val")
	})
	redbench.Bench("GET", "127.0.0.1:6379", nil, nil, func(buf []byte) []byte {
		return redbench.AppendCommand(buf, "GET", "key:string")
	})

	redbench.Bench("SET", "127.0.0.1:6380", nil, nil, func(buf []byte) []byte {
		return redbench.AppendCommand(buf, "SET", "key:string", "val")
	})
	redbench.Bench("GET", "127.0.0.1:6380", nil, nil, func(buf []byte) []byte {
		return redbench.AppendCommand(buf, "GET", "key:string")
	})
}
