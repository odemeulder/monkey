package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"demeulder.us/monkey/compiler"
	"demeulder.us/monkey/evaluator"
	"demeulder.us/monkey/lexer"
	"demeulder.us/monkey/object"
	"demeulder.us/monkey/parser"
	"demeulder.us/monkey/vm"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	// environment := object.NewEnvironment(nil)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, fmt.Sprintf("%s\n", program.String()))

		// evaluated := evaluator.Eval(program, environment)
		// if evaluated != nil {
		// 	io.WriteString(out, evaluated.Inspect())
		// 	io.WriteString(out, "\n")
		// }

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElement()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")

	}
}

func InterpretFile(fname string) {

	b, err := os.ReadFile("./programs/" + fname)
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
	}

	l := lexer.New(string(b))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(os.Stdout, p.Errors())
	}

	io.WriteString(os.Stdout, fmt.Sprintf("Program:\n%s\n", program.String()))

	environment := object.NewEnvironment(nil)
	evaluated := evaluator.Eval(program, environment)
	if evaluated != nil {
		io.WriteString(os.Stdout, fmt.Sprintf("Result:\n%s\n", evaluated.Inspect()))
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, e := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s\n", e))
	}
}
