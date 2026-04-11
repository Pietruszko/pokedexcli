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
	"math/rand"
)

type config struct {
	next *string
	previous *string
	cache *pokecache.Cache
	pokedex map[string]pokemon
}

type cliCommand struct {
	name string
	description string
	callback func(*config, []string) error
}

type locationResponse struct {
	Next *string `json:"next"`
	Previous *string `json:"previous"`
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

type exploreResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type pokemon struct {
	Name string `json:"name"`
	BaseExperience int `json:"base_experience"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	commands := map[string]cliCommand{}

	rand.Seed(time.Now().UnixNano())

	cfg := &config{
		cache: pokecache.NewCache(5 * time.Second),
		pokedex: make(map[string]pokemon),
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

	commands["explore"] = cliCommand{
		name:					"explore",
		description: 	"Explore a location area",
		callback: 		commandExplore,
	}

	commands["catch"] = cliCommand{
		name:					"catch",
		description:	"Catch a pokemon",
		callback:			commandCatch,
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

		err := cmd.callback(cfg, words[1:])
		if err != nil {
			fmt.Println(err)
		}
	}
}

func commandExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands map[string]cliCommand) func(*config, []string) error {
	return func(cfg *config, args []string) error {
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

func fetchExplore(cfg *config, url string) (exploreResponse, error) {
	if data, ok := cfg.cache.Get(url); ok {
		var cached exploreResponse
		err := json.Unmarshal(data, &cached)
		if err != nil {
			return exploreResponse{}, err
		}
		fmt.Println("(cache hit)")
		return cached, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return exploreResponse{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return exploreResponse{}, err
	}

	cfg.cache.Add(url, body)

	var data exploreResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return exploreResponse{}, err
	}

	return data, nil
}

func fetchPokemon(cfg *config, name string) (pokemon, error) {
	url := "https://pokeapi.co/api/v2/pokemon/" + name

	// cache check
	if data, ok := cfg.cache.Get(url); ok {
		var p pokemon 
		err := json.Unmarshal(data, &p)
		if err != nil {
			return pokemon{}, err
		}
		fmt.Println("(cache hit)")
		return p, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return pokemon{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return pokemon{}, err
	}

	cfg.cache.Add(url, body)

	var p pokemon
	err = json.Unmarshal(body, &p)
	if err != nil {
		return pokemon{}, err
	}

	return p, nil
}

func commandExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("please provide a location area")
		return nil
	}

	area := args[0]

	fmt.Println("Exploring %s...\n", area)

	url := "https://pokeapi.co/api/v2/location-area/" + area

	data, err := fetchExplore(cfg, url)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	
	for _, p := range data.PokemonEncounters {
		fmt.Printf(" - %s\n,", p.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("please provide a pokemon name")
		return nil
	}

	name := args[0]

	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	p, err := fetchPokemon(cfg, name)
	if err != nil {
		return err
	}

	// catch chance
	chance := rand.Intn(100)
	threshold := 50 - (p.BaseExperience / 10)

	if chance > threshold {
		fmt.Printf("%s was caught!\n", name)
		cfg.pokedex[name] = p
	} else {
		fmt.Printf("%s escaped!\n", name)
	}

	return nil
}

func commandMap(cfg *config, args []string) error {
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

func commandMapBack(cfg *config, args []string) error {
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
