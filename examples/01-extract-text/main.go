package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/derickschaefer/portabletext"
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: extract-text <file.json>")
		fmt.Println("Extracts plain text from Portable Text JSON")
		os.Exit(1)
	}

	filename := flag.Arg(0)

	// Read file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Decode Portable Text
	doc, err := portabletext.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	// Extract and print text
	for i, node := range doc {
		if node.IsBlock() {
			style := node.GetStyle()
			text := node.GetText()

			// Add some formatting based on style
			switch style {
			case "h1":
				fmt.Printf("\n# %s\n\n", text)
			case "h2":
				fmt.Printf("\n## %s\n\n", text)
			case "h3":
				fmt.Printf("\n### %s\n\n", text)
			default:
				if node.ListItem != nil {
					fmt.Printf("  • %s\n", text)
				} else {
					fmt.Printf("%s\n", text)
				}
			}
		} else {
			fmt.Printf("[Custom node %d: %s]\n", i, node.Type)
		}
	}
}
