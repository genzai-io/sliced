package document

import (
	"github.com/valyala/fastjson"
)

// Converts from a JSON representation to it's equivalent Protobuf representation.
func (mt *MessageType) JSONtoPBUF(scanner *fastjson.Parser, buf []byte) ([]byte, error) {
	//value := scanner.ParseBytes(buf)
	//value.Object()
	return nil, nil
}

func (mt *MessageType) PBUFtoJSON(buf []byte) {
	// Iterate through
}
