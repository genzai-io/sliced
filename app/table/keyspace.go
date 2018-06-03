package table

func ParsePrimaryKey(arg []byte) Key {
	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case ':':
			return Key2{StringKey(arg[:i]), ParseKeyBytes(arg[i+1:])}
		}
	}
	return ParseKeyBytes(arg)
}

func ParseTri(arg []byte) (string, string, string) {
	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case ':':
			return "", "", ""
			//return Key2{StringKey(arg[:i]), ParseKeyBytes(arg[i+1:])}
		}
	}
	return "", "", ""
}
