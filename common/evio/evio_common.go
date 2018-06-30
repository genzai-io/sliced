package evio

import "github.com/genzai-io/sliced/common/fastjson"

type LoopHelpers struct {
}

type JSON struct {
	Scanner *fastjson.Scanner
	Parser  *fastjson.Parser
}
