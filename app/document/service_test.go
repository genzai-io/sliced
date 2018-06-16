package document

import (
	"fmt"
	"testing"

	"github.com/genzai-io/sliced/proto/store"
)

func Test_ProtoService(t *testing.T) {
	err := Service.Start()
	if err != nil {
		t.Fatal(err)
	}

	//fileDescriptor, descriptor := descriptor.ForMessage(&store.Topic{})

	gz, _ := (&store.Topic{}).Descriptor()

	p, err := Service.AddFile(gz)
	if err != nil {
		t.Fatal(err)
	}

	topicProto := p.Messages["Topic"]
	_ = topicProto

	//fmt.Println(topicProto)

	topic := &store.Topic{}
	topic.Name = "allocate"
	topic.Id = -1002
	topic.Key = &store.Projection{
		Codec: store.Codec_JSON,
		Names: []string{"name", "id"},
	}
	topic.QueueID = -1

	data, _ := topic.Marshal()



	topicProto.PBUFKeyIterator(data, topicProto.FieldByNumber, func(entry *PBUFKey) bool {
		if entry.Field != nil {
			fmt.Println(fmt.Sprintf("Field Name: %s\nField Type:%s\n", entry.Field.Name, entry.Field.ProtobufType))
		}
		fmt.Println(entry.Key)
		return true
	})


	document := ToDocument(data)
	fmt.Println(document.Type())

	//nameExp := "name"

	//segmentProto := p.Messages["Node"]
	//fmt.Println(segmentProto)

	//data, err := proto.Marshal(fileDescriptor)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(data)
	//
	//fmt.Println(fileDescriptor)
	//fmt.Println(descriptor)
}
