package compiler

import (
	"fmt"

	"demeulder.us/monkey/ast"

	"demeulder.us/monkey/code"
	"demeulder.us/monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object

	lastInstruction EmittedInstruction
	prevInstruction EmittedInstruction

	symbolTable *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},

		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},

		symbolTable: NewSymbolTable(),
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

func (c *Compiler) Compile(node ast.Node) error {

	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreater)
		case "<":
			c.emit(code.OpLess)
		case ">=":
			c.emit(code.OpGreatorEqual)
		case "<=":
			c.emit(code.OpLessEqual)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknwown operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		addr := c.addConstant(integer)
		c.emit(code.OpConstant, addr)
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.StringLiteral:
		s := &object.String{Value: node.Value}
		addr := c.addConstant(s)
		c.emit(code.OpConstant, addr)
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unknwown prefix operator %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.instructions)
		newInstruction := code.Make(code.OpJumpNotTruthy, afterConsequencePos)
		c.replaceInstruction(jumpNotTruthyPos, newInstruction)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}
		afterAlternativePos := len(c.instructions)
		newInstruction = code.Make(code.OpJump, afterAlternativePos)
		c.replaceInstruction(jumpPos, newInstruction)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
	case *ast.ArrayLiteral:
		for _, e := range node.Items {
			err := c.Compile(e)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Items))
	}
	return nil
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	c.setLastInstruction(op, posNewInstruction)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	prev := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.lastInstruction = last
	c.prevInstruction = prev
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.prevInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstr []byte) {
	for i := 0; i < len(newInstr); i++ {
		c.instructions[pos+i] = newInstr[i]
	}
}
