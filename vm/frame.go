package vm

import (
	"demeulder.us/monkey/code"
	"demeulder.us/monkey/object"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
}

func NewFrame(cf *object.CompiledFunction) *Frame {
	return &Frame{fn: cf, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
