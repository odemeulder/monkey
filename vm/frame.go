package vm

import (
	"demeulder.us/monkey/code"
	"demeulder.us/monkey/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	BasePointer int
}

func NewFrame(c *object.Closure, basePointer int) *Frame {
	return &Frame{cl: c, ip: -1, BasePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
