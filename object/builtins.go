package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{"len", &Builtin{Fn: monkeyLen}},
	{"puts", &Builtin{Fn: monkeyPuts}},
	{"first", &Builtin{Fn: monkeyFirst}},
	{"last", &Builtin{Fn: monkeyLast}},
	{"rest", &Builtin{Fn: monkeyRest}},
	{"push", &Builtin{Fn: monkeyPush}},
}

func monkeyLen(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *Array:
		return &Integer{Value: int64(len(arg.Items))}
	case *String:
		return &Integer{Value: int64(len(arg.Value))}
	default:
		return newError("argument to `len` not supported, got %s",
			args[0].Type())
	}
}

func monkeyFirst(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *Array:
		if len(arg.Items) == 0 {
			return nil
		}
		return arg.Items[0]
	case *String:
		if len(arg.Value) == 0 {
			return nil
		}
		return &String{Value: string(arg.Value[0])}
	default:
		return newError("argument to `first` must be ARRAY, got %s",
			args[0].Type())
	}
}

func monkeyLast(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *Array:
		if len(arg.Items) == 0 {
			return nil
		}
		return arg.Items[len(arg.Items)-1]
	case *String:
		if len(arg.Value) == 0 {
			return nil
		}
		return &String{Value: string(arg.Value[len(arg.Value)-1])}
	default:
		return newError("argument to `last` must be ARRAY, got %s",
			args[0].Type())
	}
}

func monkeyRest(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *Array:
		length := len(arg.Items)
		if length == 0 {
			return nil
		}
		rest := make([]Object, length-1)
		sl := arg.Items[1:length]
		copy(rest, sl)
		return &Array{Items: rest}
	default:
		return newError("argument to `rest` must be ARRAY, got %s",
			args[0].Type())
	}
}

func monkeyPush(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}
	switch arg := args[0].(type) {
	case *Array:
		length := len(arg.Items)
		rest := make([]Object, 0, length)
		for _, o := range arg.Items {
			rest = append(rest, o)
		}
		rest = append(rest, args[1])
		return &Array{Items: rest}
	default:
		return newError("argument to `push` must be ARRAY, got %s",
			args[0].Type())
	}
}

func monkeyPuts(args ...Object) Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return nil
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}
