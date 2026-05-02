package obfusicate

const obMask = "*"

// ConcealPrefix replaces all but the last revealChars characters of a string.
func ConcealPrefix(s string, revealChars int) string {
	if len(s) <= revealChars {
		return repeat(obMask, len(s))
	}

	return repeat(obMask, len(s)-revealChars) + s[len(s)-revealChars:]
}

func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	if count == 1 {
		return s
	}

	var r string
	for i := 0; i < count; i++ {
		r += s
	}

	return r
}
