package table

import "strconv"

func (dbi *ValueItem) AppendValue(buf []byte) []byte {
	return buf
}

func appendArray(buf []byte, count int) []byte {
	buf = append(buf, '*')
	buf = append(buf, strconv.FormatInt(int64(count), 10)...)
	buf = append(buf, '\r', '\n')
	return buf
}

func appendBulkString(buf []byte, s string) []byte {
	buf = append(buf, '$')
	buf = append(buf, strconv.FormatInt(int64(len(s)), 10)...)
	buf = append(buf, '\r', '\n')
	buf = append(buf, s...)
	buf = append(buf, '\r', '\n')
	return buf
}

// writeSetTo writes an value as a single SET record to the a bufio Writer.
func (dbi *ValueItem) writeSetTo(buf []byte) []byte {
	if dbi.Expires > 0 {
		ex := dbi.Expires
		buf = appendArray(buf, 5)
		buf = appendBulkString(buf, "SET")
		//buf = appendBulkString(buf, dbi.Extract)
		buf = appendBulkString(buf, string(dbi.AppendValue(nil)))
		buf = appendBulkString(buf, "EX")
		buf = appendBulkString(buf, strconv.FormatUint(uint64(ex), 10))
	} else {
		buf = appendArray(buf, 3)
		buf = appendBulkString(buf, "SET")
		//buf = appendBulkString(buf, dbi.Extract)
		buf = appendBulkString(buf, string(dbi.AppendValue(nil)))
	}
	return buf
}

// writeSetTo writes an value as a single DEL record to the a bufio Writer.
func (dbi *ValueItem) writeDeleteTo(buf []byte) []byte {
	buf = appendArray(buf, 2)
	buf = appendBulkString(buf, "DEL")
	//buf = appendBulkString(buf, dbi.Extract)
	return buf
}
