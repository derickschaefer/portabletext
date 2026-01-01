package main

import (
	"fmt"
	"log"

	"github.com/derickschaefer/portabletext"
)

func main() {
	// Build a document programmatically
	doc := portabletext.Document{
		// Title
		*portabletext.NewBlock("h1").
			AddSpan("My Blog Post"),

		// Introduction paragraph
		*portabletext.NewBlock("normal").
			AddSpan("This is an introduction with ").
			AddSpan("bold text", "strong").
			AddSpan(" and ").
			AddSpan("italic text", "em").
			AddSpan("."),

		// Subheading
		*portabletext.NewBlock("h2").
			AddSpan("Key Points"),

		// Bullet list
		*portabletext.NewBlock("normal").
			AddSpan("First important point"),

		*portabletext.NewBlock("normal").
			AddSpan("Second important point"),
	}

	// Paragraph with link (build separately)
	linkBlock := portabletext.NewBlock("normal")
	linkBlock.AddSpan("For more info, visit ").
		AddSpan("our website", "link1").
		AddSpan(".")
	linkBlock.AddMarkDef("link1", "link", map[string]any{
		"href": "https://example.com",
	})
	doc = append(doc, *linkBlock)

	// Custom node (build separately)
	custom := portabletext.NewNode("callout")
	custom.Raw["text"] = "This is a custom callout box!"
	custom.Raw["variant"] = "info"
	doc = append(doc, *custom)

	// Set list items manually for bullets
	listItem := "bullet"
	doc[3].ListItem = &listItem
	doc[4].ListItem = &listItem

	// Validate
	if errs := portabletext.Validate(doc); len(errs) > 0 {
		log.Println("Validation warnings:")
		for _, err := range errs {
			log.Printf("  - %v", err)
		}
	}

	// Encode to JSON
	output, err := portabletext.EncodeString(doc)
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	fmt.Println(output)
}
