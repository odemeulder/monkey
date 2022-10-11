package vm

import (
	"fmt"

	"demeulder.us/monkey/code"
	"demeulder.us/monkey/compiler"
	"demeulder.us/monkey/object"
)

const StackSize = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

type VirtualMachine struct {
	instructions code.Instructions
	constants    []object.Object

	stack []object.Object
	sp    int // always point to the next element in the stack, top of the stack is stack[sp-1]
}

func New(bc *compiler.Bytecode) *VirtualMachine {
	return &VirtualMachine{
		instructions: bc.Instructions,
		constants:    bc.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VirtualMachine) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
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
			constindex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
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
		}
	}
	return nil
}

func (vm *VirtualMachine) executeBangExpression(op code.Opcode) error {
	operand := vm.pop()
	if operand.Type() == object.BOOLEAN_OBJ {
		return vm.push(&object.Boolean{Value: !operand.(*object.Boolean).Value})
	} else if operand.Type() == object.NULL_OBJ {
		return vm.push(&object.Boolean{Value: false})
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
