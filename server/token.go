package server

// isReservedFielName checks the given field name against reserved keywords.
func isReservedFieldName(field string) bool {
	switch field {
	case "z", "lat", "lon":
		return true
	}
	return false
}

// returns the first argument from the given string list and removes it from that.
func tokenval(vs []string) (nvs []string, token string, ok bool) {
	if len(vs) > 0 {
		token = vs[0]
		nvs = vs[1:]
		ok = true
	}
	return
}

// returns the first argument from the given string list with bytes format and removes it from that.
func tokenvalbytes(vs []string) (nvs []string, token []byte, ok bool) {
	if len(vs) > 0 {
		token = []byte(vs[0])
		nvs = vs[1:]
		ok = true
	}
	return
}

// compares the given byte arrays vs the given string
func lcb(s1 []byte, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		ch := s1[i]
		if ch >= 'A' && ch <= 'Z' {
			if ch+32 != s2[i] {
				return false
			}
		} else if ch != s2[i] {
			return false
		}
	}
	return true
}
