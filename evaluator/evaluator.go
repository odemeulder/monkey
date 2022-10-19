package evaluator

import (
	"fmt"

	"demeulder.us/monkey/ast"
	"demeulder.us/monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBooleanToBooleanOjbect(node.Value)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalConditionalExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node.Value, env)
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Environment: env}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ArrayLiteral:
		items := evalExpressions(node.Items, env)
		return &object.Array{Items: items}
	case *ast.IndexExpression:
		array := Eval(node.Left, env)
		if isError(array) {
			return array
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(array, index)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

func evalExpressions(arguments []ast.Expression, env *object.Environment) []object.Object {
	retVal := []object.Object{}
	for _, p := range arguments {
		val := Eval(p, env)
		if isError(val) {
			return []object.Object{val}
		}
		retVal = append(retVal, val)
	}
	return retVal
}

func isTruthy(obj object.Object) bool {
	var ret bool
	switch obj.Type() {
	case object.NULL_OBJ:
		ret = false
	case object.BOOLEAN_OBJ:
		ret = obj.(*object.Boolean).Value
	default:
		ret = true
	}
	return ret
}

func nativeBooleanToBooleanOjbect(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(node *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range node.Statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalConditionalExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return evalBlockStatement(ie.Consequence, env)
	} else {
		if ie.Alternative != nil {
			return evalBlockStatement(ie.Alternative, env)
		}
		return NULL
	}
}

func evalBlockStatement(node *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range node.Statements {
		result = Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorEspression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalInfixIntegerExpression(operator, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalInfixBooleanExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalInfixStringExpression(operator, left, right)
	}
	return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func evalInfixBooleanExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Boolean).Value
	rightVal := right.(*object.Boolean).Value
	switch operator {
	case "==":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalInfixStringExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalInfixIntegerExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
	case "==":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalBangOperatorEspression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalIdentifier(identifier string, env *object.Environment) object.Object {
	if obj, ok := env.Get(identifier); ok {
		return obj
	}
	if obj, ok := builtins[identifier]; ok {
		return obj
	}
	return newError("identifier not found: %s", identifier)
}

func applyFunction(function object.Object, args []object.Object, env *object.Environment) object.Object {
	switch fn := function.(type) {
	case *object.Function:
		extendedEnv := extendEnvironment(args, fn)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if result := fn.Fn(args...); result != nil {
			return result
		}
		return NULL
	default:
		return newError("not a function: %s\n", fn.Type())
	}
}

func extendEnvironment(args []object.Object, fn *object.Function) *object.Environment {
	extEnv := object.NewEnvironment(fn.Environment)
	for i, p := range fn.Parameters {
		extEnv.Set(p.Value, args[i])
	}
	return extEnv
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Items) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Items[idx]
}

func evalHashIndexExpression(hash object.Object, key object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	hashableKey, ok := key.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", key.Type())
	}
	pair, ok := hashObject.Pairs[hashableKey.HashKey()]
	if !ok {
		return NULL
	}
	return pair.Value
}

func evalHashLiteral(hl *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for k, v := range hl.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return key
		}
		hashableKey, ok := key.(object.Hashable)
		if !ok {
			return newError("key for hash does is not hashable")
		}
		value := Eval(v, env)
		if isError(value) {
			return value
		}
		pairs[hashableKey.HashKey()] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}
