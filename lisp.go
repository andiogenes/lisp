package lisp

import (
	"bufio"
	"fmt"
	"github.com/andiogenes/lisp-101/parser"
	"log"
	"os"
)

// Scope defines environment, where code evaluates
// It contains:
// field "definitions" - map, where program seeks for values, which associated with some names,
// field Outer - pointer to outer environment, where program seeks for definitions, when find nothing in current.
type Scope struct {
	definitions map[string]interface{}
	Outer       *Scope
}

// Global is global environment, you can access to it in common part of code
var Global = &Scope{definitions: map[string]interface{}{
	"+": func(l parser.SXList) interface{} {
		acc := 0
		for _, v := range l {
			acc += v.(int)
		}
		return acc
	},
	"-": func(l parser.SXList) interface{} {
		if len(l) == 1 {
			return -(l[0].(int))
		}
		s := l[0].(int)
		for _, v := range l[1:] {
			s -= v.(int)
		}
		return s
	},
	"*": func(l parser.SXList) interface{} {
		acc := 1
		for _, v := range l {
			acc *= v.(int)
		}
		return acc
	},
	"eq?": func(l parser.SXList) interface{} {
		return l[0] == l[1]
	},
	">": func(l parser.SXList) interface{} {
		return l[0].(int) > l[1].(int)
	},
	">=": func(l parser.SXList) interface{} {
		return l[0].(int) >= l[1].(int)
	},
	"<": func(l parser.SXList) interface{} {
		return l[0].(int) < l[1].(int)
	},
	"<=": func(l parser.SXList) interface{} {
		return l[0].(int) <= l[1].(int)
	},
	"car": func(l parser.SXList) interface{} {
		return l[0].(parser.SXList)[0]
	},
	"cdr": func(l parser.SXList) interface{} {
		return l[0].(parser.SXList)[1:]
	},
	"append": func(l parser.SXList) interface{} {
		var constructed parser.SXList
		for _, v := range l {
			constructed = append(constructed, v.(parser.SXList)...)
		}
		return constructed
	},
	"null?": func(l parser.SXList) interface{} {
		return len(l[0].(parser.SXList)) == 0
	},
	"map": func(l parser.SXList) interface{} { // Добавить проверку на 1 параметр в ф-ии для соблюдения соглашения
		fun, list := l[0].(func(parser.SXList) interface{}), l[1].(parser.SXList)
		var reduced parser.SXList
		for ; len(list) != 0; list = list[1:] {
			reduced = append(reduced, fun(list))
		}
		return reduced
	},
	"filter": func(l parser.SXList) interface{} {
		fun, list := l[0].(func(parser.SXList) interface{}), l[1].(parser.SXList)
		var reduced parser.SXList
		for ; len(list) != 0; list = list[1:] {
			if cond, _ := fun(list).(bool); cond == true {
				reduced = append(reduced, list[0])
			}
		}
		return reduced
	},
	"fold": func(l parser.SXList) interface{} {
		fun, reduced, list := l[0].(func(parser.SXList) interface{}), l[1].(interface{}), l[2].(parser.SXList)
		for ; len(list) != 0; list = list[1:] {
			reduced = fun(parser.SXList{reduced, list[0]})
		}
		return reduced
	},
},
	Outer: nil}

// MakeScope creates empty environment with and nests it into other
func MakeScope(outer *Scope) *Scope {
	return &Scope{make(map[string]interface{}), outer}
}

// Define associates key string with specified value in scope
func (s *Scope) Define(key string, value interface{}) {
	s.definitions[key] = value
}

// HasKey checks, does specified key belong to scope
// returns true, if key found, false if otherwise
func (s *Scope) HasKey(key string) bool {
	_, exists := s.definitions[key]
	return exists
}

// GetKey tries to get value from definitions map as key specified
// returns pair of values:
// * obtained value of some data type or nil
// * boolean, which shows existance of key/value pair in map
func (s *Scope) GetKey(key string) (value interface{}, exists bool) {
	value, exists = s.definitions[key]
	return
}

// FindInScopes tries to get value from definitions map and if
// there is no match in current scope, moves to outer scope and tries
// to do same thing there
// returns pair of values:
// * obtained value of some data type or nil
// * boolean, which shows existance of key/value pair in map
func (s *Scope) FindInScopes(key string) (value interface{}, exists bool) {
	value, exists = s.GetKey(key)
	if exists || s.Outer == nil {
		return
	}
	value, exists = s.Outer.GetKey(key)
	return
}

// REPL runs Read-Eval-Print-Loop cycle.
func REPL() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			REPL() // Сомнительно
		}
	}()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("> ")
		scanner.Scan()
		read := scanner.Text()
		if read == "quit" {
			break
		}
		evaluate := Eval(parser.Parse(read), Global)
		fmt.Println(evaluate.(parser.SXList)[0])
	}
}

// Eval evaluates specified code in specified scope
// returns result of evaluation
func Eval(exp interface{}, scope *Scope) interface{} {
	list := exp.(parser.SXList)
	var evaluated parser.SXList

	for _, v := range list {
		evaluated = append(evaluated, eval(v, scope))
	}

	return evaluated
}

func eval(exp interface{}, scope *Scope) interface{} {
	switch exp.(type) {
	case int, bool:
		return exp
	case string:
		r := exp.(string)
		if r[0] == '"' { // Добавить проверку на кавычку в конце слова
			return r[1 : len(r)-1]
		}
		if r[0] == '\'' {
			return r
		}
		reduced, found := scope.FindInScopes(r)
		if found {
			return reduced
		}
		panic(fmt.Errorf("Undefined: %v", r)) // Заменить на лучший обработчик
	case parser.SXList:
		head, tail := exp.(parser.SXList)[0], exp.(parser.SXList)[1:]
		if form, success := head.(string); success {
			switch form {
			case "define":
				id, isString := exp.(parser.SXList)[1].(string)
				value := eval(exp.(parser.SXList)[2], scope)
				if isString {
					scope.Define(id, value)
				} else {
					reduced := eval(id, scope).(string)
					scope.Define(reduced, value)
				}
				return nil // Пусть возвращает nil или некий ErrorRaiser для избежания undefined behaviour?
			case "lambda", "λ":
				args, body := exp.(parser.SXList)[1].(parser.SXList), exp.(parser.SXList)[2]
				localScope := MakeScope(scope)

				return func(l parser.SXList) interface{} {
					for i, v := range args {
						localScope.Define(v.(string), l[i])
					}

					return eval(body, localScope)
				}
			case "if":
				cond, isBool := exp.(parser.SXList)[1].(bool)
				if !isBool {
					cond = eval(exp.(parser.SXList)[1], scope).(bool)
				}
				if cond {
					return eval(exp.(parser.SXList)[2], scope)
				}
				return eval(exp.(parser.SXList)[3], scope)
			case "quote":
				return exp.(parser.SXList)[1]
			}
		}
		proc := eval(head, scope).(func(parser.SXList) interface{})
		var args parser.SXList
		for _, v := range tail {
			args = append(args, eval(v, scope))
		}
		return proc(args)
	}
	return nil
}
