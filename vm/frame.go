package vm

import (
	"demeulder.us/monkey/code"
	"demeulder.us/monkey/object"
)

type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	BasePointer int
}

func NewFrame(cf *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{fn: cf, ip: -1, BasePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
