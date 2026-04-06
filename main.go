package main

import (
	"fmt"
	"bufio"
	"os"
	"encoding/json"
	"net/http"
	"github.com/Pietruszko/pokedexcli/internal/pokecache"
	"time"
	"io"
)

type config struct {
	next *string
	previous *string
	cache *pokecache.Cache
}

type cliCommand struct {
	name string
	description string
	callback func(*config) error
}

type locationResponse struct {
	Next *string `json:"next"`
	Previous *string `json:"previous"`
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	commands := map[string]cliCommand{}

	cfg := &config{
		cache: pokecache.NewCache(5 * time.Second),
	}

	commands["exit"] = cliCommand{
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	}

	commands["map"] = cliCommand{
		name:        "map",
		description: "Displays next 20 locations",
		callback:    commandMap,
	}

	commands["mapb"] = cliCommand{
		name:        "mapb",
		description: "Displays previous 20 locations",
		callback:    commandMapBack,
	}

	commands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp(commands),
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

		err := cmd.callback(cfg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands map[string]cliCommand) func(*config) error {
	return func(cfg *config) error {
		fmt.Println("Welcome to the Pokedex!")
		fmt.Println("Usage:")

		for _, cmd := range commands {
			fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		}

		return nil
	}
}

func fetchLocations(cfg *config, url string) (locationResponse, error) {
	if data, ok := cfg.cache.Get(url); ok {
		var cached locationResponse
		err := json.Unmarshal(data, &cached)
		if err != nil {
			return locationResponse{}, err
		}
		fmt.Println("(cache hit)")
		return cached, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return locationResponse{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return locationResponse{}, err
	}
	
	cfg.cache.Add(url, body)

	var data locationResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return locationResponse{}, err
	}

	return data, nil
}

func commandMap(cfg *config) error {
	url := "https://pokeapi.co/api/v2/location-area"

	if cfg.next != nil {
		url = *cfg.next
	}

	data, err := fetchLocations(cfg, url)
	if err != nil {
		return err
	}

	cfg.next = data.Next
	cfg.previous = data.Previous

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}

	return nil
}

func commandMapBack(cfg *config) error {
	if cfg.previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	data, err := fetchLocations(cfg, *cfg.previous)
	if err != nil {
		return err
	}

	cfg.next = data.Next
	cfg.previous = data.Previous

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}

	return nil
}
