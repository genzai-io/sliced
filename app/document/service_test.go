package document

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/genzai-io/sliced/app/table"
	"github.com/genzai-io/sliced/common/gjson"
	"github.com/genzai-io/sliced/proto/store"
)

var mu sync.Mutex
var srv *ProtoService
var file *ProtoFile

func createService() (*ProtoService, *ProtoFile) {
	mu.Lock()
	defer mu.Unlock()

	if srv != nil {
		return srv, file
	}
	srv = NewProtoService()
	srv.Start()

	gz, _ := (&store.Topic{}).Descriptor()

	f, err := Service.AddFile(gz)
	if err != nil {
		panic(err)
	}

	file = f

	return srv, f
}

func Test_ProtoService(t *testing.T) {
	_, file := createService()

	topicProto := file.Messages["Topic"]

	//fmt.Println(topicProto)

	topic := &store.Topic{}
	topic.Name = "allocate"
	topic.Id = -1002
	topic.Key = &store.Projection{
		Codec: store.Codec_JSON,
		Names: []string{"name", "id"},
	}
	topic.QueueID = -1

	fmt.Println(topic.Size())

	data, _ := topic.Marshal()
	
	jsonData, _ := json.Marshal(topic)
	fmt.Println(len(data))
	fmt.Println(len(jsonData))


	idField := topicProto.FieldsByName["id"]
	nameField := topicProto.FieldsByName["name"]
	queueIDField := topicProto.FieldsByName["queueID"]

	fields := []*FieldType{
		idField,
		nameField,
		queueIDField,
	}
	keysBuf := make([]table.Key, 0, 128)

	for i := 0; i < 5; i++ {
		keys, err := topicProto.PBUFGet(data, fields, keysBuf)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(keys)
	}


	//key := topicProto.PBUFProject(data, nameProjector)
	//fmt.Println(key)

	jsonKey := gjson.GetBytes(jsonData, "name")
	fmt.Println(jsonKey)

	//topicProto.PBUFKeyIterator(data, topicProto.FieldByNumber, func(entry *PBUFKey) bool {
	//	if entry.Field != nil {
	//		fmt.Println(fmt.Sprintf("Field Name: %s\nField Type:%s\n", entry.Field.Name, entry.Field.ProtobufType))
	//	}
	//	fmt.Println(entry.Key)
	//	return true
	//})


	document := ToDocument(data)
	fmt.Println(document.Type())
}

func TestPBUFGetMultiple_OutOfOrder(t *testing.T) {
	_, file := createService()

	topicProto := file.Messages["Topic"]

	topic := &store.Topic{}
	topic.Name = "allocate"
	topic.Id = 1002
	topic.Key = &store.Projection{
		Codec: store.Codec_JSON,
		Names: []string{"name", "id"},
	}
	topic.QueueID = -1

	fmt.Println(topic.Size())

	data, _ := topic.Marshal()

	jsonData, _ := json.Marshal(topic)
	fmt.Println(len(data))
	fmt.Println(len(jsonData))


	idField := topicProto.FieldsByName["id"]
	nameField := topicProto.FieldsByName["name"]
	queueIDField := topicProto.FieldsByName["queueID"]

	var keysBuf []table.Key
	keysBuf = make([]table.Key, 0, 128)

	fields := []*FieldType{
		idField,
		queueIDField,
		nameField,
	}

	keys, err := topicProto.PBUFGet(data, fields, keysBuf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(keys)
	//
	for i := 0; i < 10; i++ {
		topicProto.PBUFGet(data, fields, keysBuf)
	}
}

func BenchmarkProtobufProjector(b *testing.B) {
	_, file := createService()

	topicProto := file.Messages["Topic"]

	topic := &store.Topic{}
	topic.Name = "allocate"
	topic.Id = 1002
	topic.Key = &store.Projection{
		Codec: store.Codec_JSON,
		Names: []string{"name", "id"},
	}
	topic.QueueID = -1

	data, _ := topic.Marshal()

	// Get fields
	idField := topicProto.FieldsByName["id"]
	nameField := topicProto.FieldsByName["name"]
	queueIDField := topicProto.FieldsByName["queueID"]

	var keysBuf []table.Key
	keysBuf = make([]table.Key, 0, 128)

	fields := []*FieldType{
		idField,
		nameField,
		queueIDField,
	}

	keys, err := topicProto.PBUFGet(data, fields, keysBuf)
	if err != nil {
		b.Fatal(err)
	}
	fmt.Println(keys)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		topicProto.PBUFGet(data, fields, keysBuf)
	}
}

func BenchmarkJSONFieldProjector(b *testing.B) {
	_, file := createService()

	topicProto := file.Messages["Topic"]
	_ = topicProto

	topic := &store.Topic{}
	topic.Name = "allocate"
	topic.Id = -1002
	topic.Key = &store.Projection{
		Codec: store.Codec_JSON,
		Names: []string{"name", "id"},
	}
	topic.QueueID = -1

	jsonData, _ := json.Marshal(topic)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gjson.GetManyBytes(jsonData, "id", "name", "queueID")
	}
}