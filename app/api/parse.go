package api

import "errors"

var ErrInvalidParam = errors.New("invalid param")

//
func ParseBool(arg []byte) (bool, error) {
	switch len(arg) {
	case 0:
		return false, nil
	case 1:
		switch arg[0] {
		case 0x00:
			return false, nil
		case 0x01:
			return true, nil
		case '1', 'T', 't', 'Y', 'y':
			return true, nil

		case '0', 'F', 'f', 'N', 'n':
			return false, nil
		}
		return false, ErrInvalidParam

	case 2:
		switch arg[0] {
		case 'N', 'n':
			switch arg[1] {
			case 'O', 'o':
				return false, nil
			}
		}
		return false, ErrInvalidParam

	case 3:
		switch arg[0] {
		case 'Y', 'y':
			switch arg[1] {
			case 'E', 'e':
				switch arg[2] {
				case 'S', 's':
					return true, nil
				}
			}
		}
		return false, ErrInvalidParam
	case 4:
		switch arg[0] {
		case 'T', 't':
			switch arg[1] {
			case 'R', 'r':
				switch arg[2] {
				case 'U', 'u':
					switch arg[3] {
					case 'E', 'e':
						return true, nil
					}
				}
			}
		}
		return false, ErrInvalidParam
	case 5:
		switch arg[0] {
		case 'F', 'f':
			switch arg[1] {
			case 'A', 'a':
				switch arg[2] {
				case 'L', 'l':
					switch arg[3] {
					case 'S', 's':
						switch arg[4] {
						case 'E', 'e':
							return false, nil
						}
					}
				}
			}
		}
		return false, ErrInvalidParam
	}
	return false, ErrInvalidParam
}