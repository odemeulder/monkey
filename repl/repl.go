package repl

import (
	"bufio"
	"fmt"
	"io"

	"demeulder.us/monkey/evaluator"
	"demeulder.us/monkey/lexer"
	"demeulder.us/monkey/object"
	"demeulder.us/monkey/parser"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	environment := object.NewEnvironment(nil)

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

		evaluated := evaluator.Eval(program, environment)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, e := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s\n", e))
	}
}
