/*
Package portabletext provides parsing, validation, and manipulation of Portable Text documents.

Portable Text is a JSON-based rich text specification from Sanity.io that represents
structured content as an abstract syntax tree (AST). This package provides a complete
Go implementation with strong typing, path-aware errors, and flexible validation.

# Quick Start

Parse Portable Text from JSON:

	input := `[{"_type":"block","children":[{"_type":"span","text":"Hello"}]}]`
	doc, err := portabletext.DecodeString(input)
	if err != nil {
		log.Fatal(err)
	}

Access the parsed content:

	for _, node := range doc {
		if node.IsBlock() {
			fmt.Println(node.GetText())
		}
	}

Build documents programmatically:

	block := portabletext.NewBlock("normal").
		AddSpan("Hello ", "strong").
		AddSpan("world!")
	doc := portabletext.Document{*block}

# Core Types

The main types are:

  - Document: An ordered list of Portable Text nodes
  - Node: A block or custom object with type, style, children, and marks
  - Span: An inline text element within a block
  - MarkDef: A mark definition (e.g., link with href)

All types preserve unknown/custom fields in a Raw map for full round-trip fidelity.

# Decoding and Encoding

Decode from io.Reader or string:

	doc, err := portabletext.Decode(reader)
	doc, err := portabletext.DecodeString(jsonString)

Encode to io.Writer or string:

	err := portabletext.Encode(writer, doc)
	jsonString, err := portabletext.EncodeString(doc)

# Validation

Basic validation checks for required fields and proper structure:

	errs := portabletext.Validate(doc)
	for _, err := range errs {
		fmt.Println(err)
	}

Advanced validation with custom options:

	opts := portabletext.ValidationOptions{
		RequireKeys:      true,  // Require _key on blocks
		CheckMarkDefRefs: true,  // Verify mark references
		AllowEmptyText:   false, // Disallow empty spans
	}
	errs := portabletext.ValidateWithOptions(doc, opts)

# Traversal

Walk all nodes:

	err := portabletext.Walk(doc, func(node *portabletext.Node) error {
		fmt.Println(node.Type)
		return nil
	})

Walk with context information:

	portabletext.WalkWithContext(doc, func(n *portabletext.Node, ctx portabletext.WalkContext) error {
		fmt.Printf("Node %d: %s\n", ctx.Index, n.Type)
		return nil
	})

# Filtering and Transformation

Filter nodes by predicate:

	blocks := portabletext.Filter(doc, func(n *portabletext.Node) bool {
		return n.IsBlock()
	})

Transform nodes (returns new document):

	transformed := portabletext.Transform(doc, func(n *portabletext.Node) *portabletext.Node {
		if n.GetStyle() == "h1" {
			h2 := "h2"
			n.Style = &h2
		}
		return n
	})

# Working with Nodes

Node provides convenience methods:

	// Check type
	if node.IsBlock() { }

	// Get values with defaults
	style := node.GetStyle()        // "normal" if nil
	level := node.GetListLevel()    // 1 if nil
	text := node.GetText()          // concatenated span text

	// Build fluently
	node.AddSpan("text", "strong", "em")
	node.AddMarkDef("link1", "link", map[string]any{"href": "https://..."})

	// Clone (deep copy)
	clone := node.Clone()

# Working with Spans

Check for marks:

	if span.HasMark("strong") {
		fmt.Println("Bold text:", *span.Text)
	}

Access custom fields:

	if href, ok := markDef.Raw["href"].(string); ok {
		fmt.Println("Link:", href)
	}

# Error Handling

Errors include path information for debugging:

	doc, err := portabletext.Decode(reader)
	if err != nil {
		var pErr *portabletext.Error
		if errors.As(err, &pErr) {
			// pErr.Path shows where the error occurred
			// e.g., "[2].children[1].marks"
			fmt.Printf("Error at %s: %v\n", pErr.Path, pErr.Err)
		}
	}

Validation errors provide structured information:

	for _, err := range errs {
		if ve, ok := err.(*portabletext.ValidationError); ok {
			fmt.Printf("%s: %s\n", ve.Path, ve.Message)
			// ve.Node provides reference to the problematic node
		}
	}

# Custom Fields

Unknown fields are preserved in Raw maps:

	node.Raw["customField"] = "value"
	span.Raw["customAttr"] = 123
	markDef.Raw["target"] = "_blank"

These fields are included when encoding back to JSON.

# Thread Safety

Documents are safe for concurrent reads without synchronization.
Concurrent writes require external synchronization.
Filter() and Transform() create new documents and are safe to call concurrently.

# Examples

Extract all text:

	func ExtractText(doc portabletext.Document) string {
		var buf strings.Builder
		for _, node := range doc {
			if node.IsBlock() {
				buf.WriteString(node.GetText())
				buf.WriteString("\n")
			}
		}
		return buf.String()
	}

Find all links:

	func FindLinks(doc portabletext.Document) []string {
		var links []string
		for _, node := range doc {
			for _, md := range node.MarkDefs {
				if md.Type == "link" {
					if href, ok := md.Raw["href"].(string); ok {
						links = append(links, href)
					}
				}
			}
		}
		return links
	}

Generate table of contents:

	func GenerateTOC(doc portabletext.Document) []string {
		var toc []string
		portabletext.Walk(doc, func(n *portabletext.Node) error {
			style := n.GetStyle()
			if style == "h1" || style == "h2" {
				toc = append(toc, n.GetText())
			}
			return nil
		})
		return toc
	}

# Specification

This package implements the Portable Text specification:
https://github.com/portabletext/portabletext

Key concepts:
  - Block: Top-level content node (paragraph, heading, list item, etc.)
  - Span: Inline text within a block, optionally with marks
  - Mark: Text annotation (bold, italic, link, etc.)
  - MarkDef: Mark definition with additional data (e.g., link href)
  - Custom Objects: Extensible beyond standard types
*/
package portabletext
