package main

import (
	"flag"
	"fmt"
	"time"

	"demeulder.us/monkey/compiler"
	"demeulder.us/monkey/evaluator"
	"demeulder.us/monkey/lexer"
	"demeulder.us/monkey/object"
	"demeulder.us/monkey/parser"
	"demeulder.us/monkey/vm"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

var input = `
	let fibonacci = fn(x) {
		if (x == 0) {
			0;
		} else {
			if (x == 1) {
				1;
			} else {
				fibonacci(x - 1) + fibonacci(x - 2);
			}
		}
	}
	fibonacci(38)
`

func main() {
	flag.Parse()

	var duration time.Duration
	var result object.Object

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}
		machine := vm.New(comp.Bytecode())
		start := time.Now()
		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
		}
		duration = time.Since(start)
		result = machine.LastPoppedStackElement()
	} else {
		env := object.NewEnvironment(nil)
		start := time.Now()
		result = evaluator.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf("engine=%s, result=%s, duration=%s\n", *engine, result.Inspect(), duration)
}
