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
		fmt.Println("Usage: find-links <file.json>")
		fmt.Println("Extracts all links from Portable Text JSON")
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

	// Find all links
	linkCount := 0
	err = portabletext.Walk(doc, func(node *portabletext.Node) error {
		for _, md := range node.MarkDefs {
			if md.Type == "link" {
				linkCount++

				// Extract link details
				href, _ := md.Raw["href"].(string)
				title, _ := md.Raw["title"].(string)

				fmt.Printf("[%d] Key: %s\n", linkCount, md.Key)
				fmt.Printf("    URL: %s\n", href)
				if title != "" {
					fmt.Printf("    Title: %s\n", title)
				}

				// Find text using this link
				for _, child := range node.Children {
					if child.HasMark(md.Key) && child.Text != nil {
						fmt.Printf("    Text: %s\n", *child.Text)
					}
				}
				fmt.Println()
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Walk failed: %v", err)
	}

	if linkCount == 0 {
		fmt.Println("No links found")
	} else {
		fmt.Printf("Total links: %d\n", linkCount)
	}
}
