package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/derickschaefer/portabletext"
)

func main() {
	var downgrade bool
	var pretty bool
	flag.BoolVar(&downgrade, "downgrade", false, "Downgrade headings (h1->h2, h2->h3, etc)")
	flag.BoolVar(&pretty, "pretty", false, "Pretty-print JSON output")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: transform-headings [--downgrade] [--pretty] <file.json>")
		fmt.Println("Transforms heading levels in Portable Text JSON")
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

	// Transform headings
	transformed := portabletext.Transform(doc, func(n *portabletext.Node) *portabletext.Node {
		if !n.IsBlock() {
			return n
		}

		style := n.GetStyle()
		var newStyle string

		if downgrade {
			// Downgrade: h1 -> h2, h2 -> h3, etc.
			switch style {
			case "h1":
				newStyle = "h2"
			case "h2":
				newStyle = "h3"
			case "h3":
				newStyle = "h4"
			case "h4":
				newStyle = "h5"
			case "h5":
				newStyle = "h6"
			default:
				return n
			}
		} else {
			// Upgrade: h2 -> h1, h3 -> h2, etc.
			switch style {
			case "h6":
				newStyle = "h5"
			case "h5":
				newStyle = "h4"
			case "h4":
				newStyle = "h3"
			case "h3":
				newStyle = "h2"
			case "h2":
				newStyle = "h1"
			default:
				return n
			}
		}

		n.Style = &newStyle
		return n
	})

	// Encode to JSON
	if pretty {
		// Pretty-print with indentation
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(transformed); err != nil {
			log.Fatalf("Failed to encode: %v", err)
		}
	} else {
		// Compact output
		output, err := portabletext.EncodeString(transformed)
		if err != nil {
			log.Fatalf("Failed to encode: %v", err)
		}
		fmt.Println(output)
	}
}
