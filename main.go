package main

import (
	"fmt"
	"bufio"
	"os"
)

type cliCommand struct {
	name string
	description string
	callback func() error
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	commands := map[string]cliCommand{}

	commands["exit"] = cliCommand{
		name: "exit",
		description: "Exit the Pokedex",
		callback: commandExit,
	}

	commands["help"] = cliCommand{
		name: "help",
		description: "Displays a help message",
		callback: commandHelp(commands),
	}

	for {
		fmt.Print("Pokedex > ")

		scanner.Scan()
		text := scanner.Text()

		words := cleanInput(text)

		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		
		cmd, exists := commands[commandName]
		if !exists {
			fmt.Println("Unknown command")
			continue
		}

		err := cmd.callback()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands map[string]cliCommand) func() error {
	return func() error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Println("Usage:")

		for _, cmd := range commands {
			fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		}

		return nil
	}
}
