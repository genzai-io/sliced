package ring

import "testing"

func BenchmarkCRC16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		crc16sum("this_is_a_longish_type_of_a_key")
	}
}
