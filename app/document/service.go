package document

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/common/base58"
	"github.com/genzai-io/sliced/common/service"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

var Service *ProtoService

func init() {
	Service = NewProtoService()
}

type ProtoService struct {
	service.BaseService

	sync.RWMutex

	files map[string]*ProtoFile
}

func NewProtoService() *ProtoService {
	s := &ProtoService{
		files: make(map[string]*ProtoFile),
	}
	s.BaseService = *service.NewBaseService(moved.Logger, "proto", Service)
	return s
}

func (p *ProtoService) OnStart() error {
	return nil
}

func (p *ProtoService) OnStop() {

}

// Add a serialized and possibly gzipped instance of a FileDescriptorProto
func (s *ProtoService) AddFile(gz []byte) (*ProtoFile, error) {
	b, err := ungzip(gz)
	if err != nil {
		b = gz
		gz, err = gzipBytes(b)

		if err != nil {
			return nil, err
		}
	}

	sha := sha256.New()
	sha.Write(gz)
	gzipHash := base58.Encode(sha.Sum(nil))

	sha.Reset()
	sha.Write(b)
	hash := base58.Encode(sha256.New().Sum(nil))

	var d *descriptor.FileDescriptorProto
	d, err = unmarshalFile(b)
	if err != nil {
		return nil, err
	}

	s.Lock()
	defer s.Unlock()

	file := NewProtoFile(hash, gzipHash, d)

	existing, ok := s.files[*d.Name]
	if !ok {

	} else {
		_ = existing
	}

	s.files[gzipHash] = file
	s.files[hash] = file

	return file, nil
}

func gzipBytes(b []byte) ([]byte, error) {
	buffer := &bytes.Buffer{}
	w := gzip.NewWriter(buffer)

	defer w.Close()

	_, err := w.Write(b)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %v", err)
	}

	err = w.Flush()
	if err != nil {
		return nil, fmt.Errorf("failed to flush gzip writer: %v", err)
	}

	return buffer.Bytes(), nil
}

func ungzip(gz []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to uncompress descriptor: %v", err)
	}
	return b, nil
}

// extractFile extracts a FileDescriptorProto from a gzip'd buffer.
func extractFile(gz []byte) (*descriptor.FileDescriptorProto, error) {
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %v", err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to uncompress descriptor: %v", err)
	}

	return unmarshalFile(b)
}

//
func unmarshalFile(b []byte) (*descriptor.FileDescriptorProto, error) {
	fd := new(descriptor.FileDescriptorProto)
	if err := proto.Unmarshal(b, fd); err != nil {
		return nil, fmt.Errorf("malformed FileDescriptorProto: %v", err)
	}

	return fd, nil
}

//
func (p *ProtoService) GetFile(hashOrPath string) (*ProtoFile, bool) {
	p.RLock()
	defer p.RUnlock()

	f, ok := p.files[hashOrPath]
	return f, ok
}

//
type ProtoFile struct {
	Name string

	Hash       string
	GzipHash   string
	Descriptor *descriptor.FileDescriptorProto
	Messages   map[string]*MessageType
	Enums      map[string]*EnumType
}

func NewProtoFile(hash, gzipHash string, descriptor *descriptor.FileDescriptorProto) *ProtoFile {
	file := &ProtoFile{
		Name:       *descriptor.Name,
		Hash:       hash,
		GzipHash:   gzipHash,
		Descriptor: descriptor,
		Messages:   make(map[string]*MessageType),
		Enums:      make(map[string]*EnumType),
	}

	newProtoMap(nil, file, descriptor.MessageType, file.Messages)
	newEnumMap(nil, file, descriptor.EnumType, file.Enums)

	for _, m := range file.Messages {
		m.resolve()
	}

	return file
}
