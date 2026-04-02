package main

import "strings"

func cleanInput(text string) []string{
	words := strings.Fields(text)
	result := make([]string, 0, len(words))

	for _, word := range words {
		clean := strings.ToLower(strings.TrimSpace(word))
		if clean != "" {
			result = append(result,clean)
		}
	}
	return result
}
