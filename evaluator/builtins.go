package evaluator

import (
	"fmt"

	"demeulder.us/monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len":   object.GetBuiltinByName("len"),
	"first": &object.Builtin{Fn: builtinFirst},
	"last":  &object.Builtin{Fn: builtinLast},
	"rest":  &object.Builtin{Fn: builtinRest},
	"push":  &object.Builtin{Fn: builtinPush},
	"puts":  &object.Builtin{Fn: builtinPuts},
}

func builtinLen(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Items))}
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	default:
		return newError("argument to `len` not supported, got %s",
			args[0].Type())
	}
}

func builtinFirst(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		if len(arg.Items) == 0 {
			return NULL
		}
		return arg.Items[0]
	case *object.String:
		if len(arg.Value) == 0 {
			return NULL
		}
		return &object.String{Value: string(arg.Value[0])}
	default:
		return newError("argument to `len` not supported, got %s",
			args[0].Type())
	}
}

func builtinLast(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		if len(arg.Items) == 0 {
			return NULL
		}
		return arg.Items[len(arg.Items)-1]
	case *object.String:
		if len(arg.Value) == 0 {
			return NULL
		}
		return &object.String{Value: string(arg.Value[len(arg.Value)-1])}
	default:
		return newError("argument to `len` not supported, got %s",
			args[0].Type())
	}
}

func builtinRest(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Items)
		if length == 0 {
			return NULL
		}
		rest := make([]object.Object, length-1)
		sl := arg.Items[1:length]
		copy(rest, sl)
		return &object.Array{Items: rest}
	default:
		return newError("argument to `rest` not supported, got %s",
			args[0].Type())
	}
}

func builtinPush(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Items)
		rest := make([]object.Object, 0, length)
		for _, o := range arg.Items {
			rest = append(rest, o)
		}
		rest = append(rest, args[1])
		return &object.Array{Items: rest}
	default:
		return newError("argument to `push` not supported, got %s",
			args[0].Type())
	}
}

func builtinPuts(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return NULL
}
