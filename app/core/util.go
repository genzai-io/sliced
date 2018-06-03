package core

import (
	"encoding/binary"

	"github.com/hashicorp/raft"
)

func decodeLog(b []byte, log *raft.Log) error {
	log.Index = binary.LittleEndian.Uint64(b[0:8])
	log.Term = binary.LittleEndian.Uint64(b[8:16])
	log.Type = raft.LogType(b[16])
	log.Data = make([]byte, int(binary.LittleEndian.Uint64(b[17:25])))
	copy(log.Data, b[25:])
	return nil
}

func encodeLog(entry *raft.Log) []byte {
	n := make([]byte, 8)       // used to store uint64s
	b := make([]byte, 42, 256) // encoded message goes here

	binary.LittleEndian.PutUint64(n, entry.Index)
	b = append(b, n...)
	binary.LittleEndian.PutUint64(n, entry.Term)
	b = append(b, n...)
	b = append(b, byte(entry.Type))
	binary.LittleEndian.PutUint64(n, uint64(len(entry.Data)))
	b = append(b, n...)
	b = append(b, entry.Data...)
	return b
}

func encodeLogs(logs []*raft.Log) []byte {
	n := make([]byte, 8)       // used to store uint64s
	b := make([]byte, 42, 256) // encoded message goes here

	binary.LittleEndian.PutUint64(n, uint64(len(logs)))
	b = append(b, n...)
	for _, entry := range logs {
		binary.LittleEndian.PutUint64(n, entry.Index)
		b = append(b, n...)
		binary.LittleEndian.PutUint64(n, entry.Term)
		b = append(b, n...)
		b = append(b, byte(entry.Type))
		binary.LittleEndian.PutUint64(n, uint64(len(entry.Data)))
		b = append(b, n...)
		b = append(b, entry.Data...)
	}
	return b
}

// Converts bytes to an integer
func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Converts a uint to a byte slice
func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
