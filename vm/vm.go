package vm

import (
	"fmt"

	"demeulder.us/monkey/code"
	"demeulder.us/monkey/compiler"
	"demeulder.us/monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1025

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VirtualMachine struct {
	constants []object.Object

	frames      []*Frame
	framesIndex int

	stack   []object.Object
	sp      int // always point to the next element in the stack, top of the stack is stack[sp-1]
	globals []object.Object
}

func New(bc *compiler.Bytecode) *VirtualMachine {
	mainFn := &object.CompiledFunction{Instructions: bc.Instructions}
	mainFrame := NewFrame(mainFn)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VirtualMachine{
		constants: bc.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VirtualMachine {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VirtualMachine) Run() error {

	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreater, code.OpGreatorEqual, code.OpLess, code.OpLessEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangExpression(op)
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeNegationExpression(op)
			if err != nil {
				return err
			}
		case code.OpConstant:
			constindex := code.ReadUint16(ins[ip+1:])
			fmt.Printf("constindex: %d\n", constindex)
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constindex])
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			conditionValue := vm.pop()
			if !isTruthy(conditionValue) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			idx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[idx] = vm.pop()
		case code.OpGetGlobal:
			idx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[idx])
			if err != nil {
				return err
			}
		case code.OpArray:
			len := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			arr := vm.buildArray(vm.sp-len, vm.sp)
			vm.sp = vm.sp - len
			err := vm.push(arr)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VirtualMachine) executeBangExpression(op code.Opcode) error {
	operand := vm.pop()
	if operand.Type() == object.BOOLEAN_OBJ {
		return vm.push(&object.Boolean{Value: !operand.(*object.Boolean).Value})
	} else if operand.Type() == object.NULL_OBJ {
		return vm.push(&object.Boolean{Value: true})
	}
	return vm.push(&object.Boolean{Value: false})
}

func (vm *VirtualMachine) executeNegationExpression(op code.Opcode) error {
	operand := vm.pop()
	if operand.Type() == object.INTEGER_OBJ {
		return vm.push(&object.Integer{Value: (-1) * operand.(*object.Integer).Value})
	}
	return fmt.Errorf("unsupported negation operation: -%s", operand.Type())
}

func (vm *VirtualMachine) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(left, right, op)
	}
	if leftType == object.STRING_OBJ && rightType == object.STRING_OBJ {
		return vm.executeBinaryStringOperation(left, right, op)
	}
	if leftType == object.BOOLEAN_OBJ && rightType == object.BOOLEAN_OBJ {
		return vm.executeBinaryBooleanOperation(left, right, op)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VirtualMachine) executeBinaryIntegerOperation(left, right object.Object, op code.Opcode) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	var result int64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("Error, unknown operator")
	}
	// fmt.Printf("%d %d %d = %d\n", leftValue, op, rightValue, result)
	return vm.push(&object.Integer{Value: result})
}

func (vm *VirtualMachine) executeBinaryStringOperation(left, right object.Object, op code.Opcode) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	var result string
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	default:
		return fmt.Errorf("Error, unknown operator")
	}
	// fmt.Printf("%d %d %d = %d\n", leftValue, op, rightValue, result)
	return vm.push(&object.String{Value: result})
}

func (vm *VirtualMachine) executeBinaryBooleanOperation(left, right object.Object, op code.Opcode) error {
	return fmt.Errorf("Error, unknown operator")
}

func (vm *VirtualMachine) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerComparison(left, right, op)
	}
	if leftType == object.BOOLEAN_OBJ && rightType == object.BOOLEAN_OBJ {
		return vm.executeBinaryBooleanComparison(left, right, op)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VirtualMachine) executeBinaryIntegerComparison(left, right object.Object, op code.Opcode) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	var result bool
	switch op {
	case code.OpEqual:
		result = leftValue == rightValue
	case code.OpNotEqual:
		result = leftValue != rightValue
	case code.OpGreater:
		result = leftValue > rightValue
	case code.OpGreatorEqual:
		result = leftValue >= rightValue
	case code.OpLess:
		result = leftValue < rightValue
	case code.OpLessEqual:
		result = leftValue <= rightValue
	default:
		return fmt.Errorf("Error, unknown operator")
	}
	// fmt.Printf("%d %d %d = %d\n", leftValue, op, rightValue, result)
	return vm.push(&object.Boolean{Value: result})
}

func (vm *VirtualMachine) executeBinaryBooleanComparison(left, right object.Object, op code.Opcode) error {
	leftValue := left.(*object.Boolean).Value
	rightValue := right.(*object.Boolean).Value
	var result bool
	if op == code.OpEqual {
		result = leftValue == rightValue
	} else {
		result = !(leftValue == rightValue)
	}
	// return fmt.Errorf("Error, unknown operator")
	return vm.push(&object.Boolean{Value: result})
}

func (vm *VirtualMachine) buildArray(startIdx int, endIdx int) *object.Array {
	arr := make([]object.Object, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		arr[i-startIdx] = vm.stack[i]
	}
	return &object.Array{Items: arr}
}

func (vm *VirtualMachine) buildHash(startIdx int, endIdx int) (*object.Hash, error) {
	hash := make(map[object.HashKey]object.HashPair, endIdx-startIdx)
	for i := startIdx; i < endIdx; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		hashPair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusuable as a hash key: %s", key)
		}
		hash[hashKey.HashKey()] = hashPair
	}
	return &object.Hash{Pairs: hash}, nil
}

func (vm *VirtualMachine) executeIndexExpression(left, index object.Object) error {
	switch left.Type() {
	case object.ARRAY_OBJ:
		if index.Type() != object.INTEGER_OBJ {
			return fmt.Errorf("array index must be integer")
		}
		return vm.executeArrayIndexExpression(left.(*object.Array), index.(*object.Integer))
	case object.HASH_OBJ:
		hashable, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("hash index must be hashable")
		}

		return vm.executeHashIndexExpression(left.(*object.Hash), hashable)
	default:
		return fmt.Errorf("invalid index expression")
	}
}

func (vm *VirtualMachine) executeArrayIndexExpression(arr *object.Array, index *object.Integer) error {
	i := index.Value
	max := int64(len(arr.Items) - 1)
	if i < 0 || i > max {
		return vm.push(Null)
	}
	return vm.push(arr.Items[i])
}

func (vm *VirtualMachine) executeHashIndexExpression(hash *object.Hash, key object.Hashable) error {
	pair, ok := hash.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pair.Value)
}

func (vm *VirtualMachine) push(obj object.Object) error {
	if vm.sp > StackSize {
		return fmt.Errorf("Stack overflow")
	}
	vm.stack[vm.sp] = obj
	vm.sp += 1
	return nil
}

func (vm *VirtualMachine) pop() object.Object {
	rv := vm.stack[vm.sp-1]
	vm.sp -= 1
	return rv
}

func (vm *VirtualMachine) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VirtualMachine) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VirtualMachine) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VirtualMachine) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VirtualMachine) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}
