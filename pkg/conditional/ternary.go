package conditional

// Ternary : ternary operator (becareful when using pointer values that are nil ...crash)
func Ternary[t any](condition bool, val1 t, val2 t) t {
	if condition {
		return val1
	}
	return val2
}
