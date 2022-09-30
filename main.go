package main

import (
	"fmt"
	"os"
	"os/user"

	"demeulder.us/monkey/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)

	if len(os.Args) > 1 {
		repl.InterpretFile(os.Args[1])
	} else {
		fmt.Printf("Feel free to type in commands\n")

		repl.Start(os.Stdin, os.Stdout)
	}

}
