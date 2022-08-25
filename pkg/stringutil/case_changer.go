package stringutil

import (
	"unicode"
	"unicode/utf8"
)

type buffer struct {
	r         []byte
	runeBytes [utf8.UTFMax]byte
}

func (b *buffer) write(r rune) {
	// if inside utf 8 range
	if r < utf8.RuneSelf {
		b.r = append(b.r, byte(r))
		return
	}

	n := utf8.EncodeRune(b.runeBytes[0:], r)
	b.r = append(b.r, b.runeBytes[0:n]...)
}

func (b *buffer) underscoreWithIndent() {
	// guard check to not underscore the first character
	if len(b.r) > 0 {
		b.r = append(b.r, '_')
	}
}

// Underscore : take a camel cased string input and turn it to snakecased
//
// Example :
//			value := Underscore("someThingLong")
//			fmt.Println(value) // some_thing_long
func Underscore(s string) string {
	b := buffer{
		r: make([]byte, 0, len(s)),
	}
	var m rune
	var w bool

	for _, ch := range s {

		// if lower , underscore the character
		// if not at th beginning we can underscore and indent
		if unicode.IsUpper(ch) {
			if m != 0 {
				if !w {
					b.underscoreWithIndent()
					w = true
				}
				b.write(m)
			}
			m = unicode.ToLower(ch)
		} else {
			if m != 0 {
				b.underscoreWithIndent()
				b.write(m)
				m = 0
				w = false
			}
			b.write(ch)
		}
	}
	if m != 0 {
		if !w {
			b.underscoreWithIndent()
		}
		b.write(m)
	}
	return string(b.r)
}
