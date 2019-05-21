package parser

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	round = iota
	square
	none
)

// SXList represents a list of S-expressions
type SXList []interface{}

func makeList(list *SXList, symbols []string, pos int, parenthType int) int {
	var i int
loop:
	for i = pos; i < len(symbols); i++ {
		switch symbols[i] {
		case "(", "[":
			var nested SXList
			if symbols[i] == "(" { // begin Плохо
				i = makeList(&nested, symbols, i+1, round)
			} else {
				i = makeList(&nested, symbols, i+1, square)
			} // end Плохо
			*list = append(*list, nested)
		case "'(", "'[": // Можно последовать KISS и объединить два кейса?
			var nested SXList
			if symbols[i] == "'(" { // begin Плохо
				i = makeList(&nested, symbols, i+1, round)
			} else {
				i = makeList(&nested, symbols, i+1, square)
			} // end Плохо
			quoted := SXList{"quote", nested}
			*list = append(*list, quoted)
		case ")", "]":
			if parenthType == none {
				panic(fmt.Errorf("Unexpected %s", symbols[i]))
			}
			var cond int // begin Очень плохо
			if symbols[i] == ")" {
				cond = round
			} else {
				cond = square
			}
			if parenthType != cond {
				panic(fmt.Errorf("Unexpected %s", symbols[i]))
			} // end Очень плохо
			break loop
		case "#t", "#true":
			*list = append(*list, true)
		case "#f", "#false":
			*list = append(*list, false)
		default:
			if s, err := strconv.Atoi(symbols[i]); err == nil {
				*list = append(*list, s)
			} else {
				*list = append(*list, symbols[i])
			}
		}
	}
	return i
}

// Parse is a function, which take a string and represents it as S-expressions list
func Parse(expr string) SXList {
	expr = strings.Replace(strings.Replace(strings.Replace(expr, "(", " ( ", -1), ")", " ) ", -1), "' (", "'(", -1)
	expr = strings.Replace(strings.Replace(strings.Replace(expr, "[", " [ ", -1), "]", " ] ", -1), "' [", "'[", -1) // Плохо
	symbols := strings.Split(expr, " ")

	for i := 0; i < len(symbols); i++ {
		if symbols[i] == "" {
			symbols = append(symbols[:i], symbols[i+1:]...)
			i = 0
			continue
		}
	}

	var representation SXList

	makeList(&representation, symbols, 0, none)

	return representation
}
