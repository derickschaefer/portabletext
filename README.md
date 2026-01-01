# portabletext

[![Go Reference](https://pkg.go.dev/badge/github.com/derickschaefer/portabletext.svg)](https://pkg.go.dev/github.com/derickschaefer/portabletext)
[![Go Report Card](https://goreportcard.com/badge/github.com/derickschaefer/portabletext)](https://goreportcard.com/report/github.com/derickschaefer/portabletext)
[![GitHub release](https://img.shields.io/github/release/derickschaefer/portabletext.svg)](https://github.com/derickschaefer/portabletext/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for parsing, validating, and manipulating [Portable Text](https://portabletext.org/) - the JSON-based rich text specification from Sanity.io.

## Features

- **Complete Portable Text Support**: Handles blocks, spans, marks, mark definitions, and custom objects
- **Preserves Unknown Fields**: Captures custom fields in a `Raw` map for full round-trip fidelity
- **Type-Safe API**: Strongly-typed structs with pointer fields for proper null handling
- **Path-Aware Errors**: Detailed error messages showing exactly where parsing/validation failed
- **Flexible Validation**: Optional validation with configurable rules
- **Zero Dependencies**: Only uses Go standard library
- **Immutable Operations**: Transform and filter operations return new documents
- **Thread-Safe Reads**: Safe for concurrent reads without synchronization

## Installation

```bash
go get github.com/derickschaefer/portabletext
```

## Quick Start

```go
package main

import (
    "fmt"
    "strings"
    
    "github.com/derickschaefer/portabletext"
)

func main() {
    // Parse Portable Text JSON
    input := `[
        {
            "_type": "block",
            "_key": "block1",
            "style": "h1",
            "children": [
                {
                    "_type": "span",
                    "text": "Hello World",
                    "marks": []
                }
            ],
            "markDefs": []
        }
    ]`
    
    doc, err := portabletext.DecodeString(input)
    if err != nil {
        panic(err)
    }
    
    // Access the data
    for _, node := range doc {
        if node.IsBlock() {
            fmt.Printf("Style: %s\n", node.GetStyle())
            fmt.Printf("Text: %s\n", node.GetText())
        }
    }
    
    // Validate
    if errs := portabletext.Validate(doc); len(errs) > 0 {
        for _, err := range errs {
            fmt.Println(err)
        }
    }
}
```

## Usage

### Decoding

```go
// From io.Reader
doc, err := portabletext.Decode(reader)

// From string
doc, err := portabletext.DecodeString(jsonString)
```

### Encoding

```go
// To io.Writer
err := portabletext.Encode(writer, doc)

// To string
jsonString, err := portabletext.EncodeString(doc)
```

### Building Documents Programmatically

```go
// Create a simple block
block := portabletext.NewBlock("normal").
    AddSpan("Hello ", "strong").
    AddSpan("world!")

// Create a custom node
customNode := portabletext.NewNode("myCustomType")
customNode.Raw["customField"] = "value"

// Build a document
doc := portabletext.Document{*block, *customNode}
```

### Validation

```go
// Basic validation
errs := portabletext.Validate(doc)

// Advanced validation with options
opts := portabletext.ValidationOptions{
    RequireKeys:      true,  // Require _key on all blocks
    CheckMarkDefRefs: true,  // Verify marks exist in markDefs
    AllowEmptyText:   false, // Disallow empty text in spans
}
errs := portabletext.ValidateWithOptions(doc, opts)

// Check for specific errors
for _, err := range errs {
    if ve, ok := err.(*portabletext.ValidationError); ok {
        fmt.Printf("Error at %s: %s\n", ve.Path, ve.Message)
    }
}
```

### Walking and Traversing

```go
// Simple walk
err := portabletext.Walk(doc, func(node *portabletext.Node) error {
    fmt.Println(node.Type)
    return nil
})

// Walk with context
err := portabletext.WalkWithContext(doc, func(node *portabletext.Node, ctx portabletext.WalkContext) error {
    fmt.Printf("Node %d (block #%d): %s\n", ctx.Index, ctx.BlockCount, node.Type)
    return nil
})
```

### Filtering and Transforming

```go
// Filter: keep only blocks
blocks := portabletext.Filter(doc, func(n *portabletext.Node) bool {
    return n.IsBlock()
})

// Transform: change all h1 to h2
transformed := portabletext.Transform(doc, func(n *portabletext.Node) *portabletext.Node {
    if n.GetStyle() == "h1" {
        h2 := "h2"
        n.Style = &h2
    }
    return n
})
```

### Working with Nodes

```go
// Check node type
if node.IsBlock() {
    // It's a block
}

// Get style with default
style := node.GetStyle() // returns "normal" if nil

// Get all text from a block
text := node.GetText()

// Get list level with default
level := node.GetListLevel() // returns 1 if nil

// Clone a node (deep copy)
clone := node.Clone()

// Add spans and marks
node.AddSpan("Click ", "strong")
node.AddMarkDef("link1", "link", map[string]any{
    "href": "https://example.com",
})
```

### Working with Spans

```go
for _, span := range node.Children {
    // Check for specific mark
    if span.HasMark("strong") {
        fmt.Println("Bold text:", *span.Text)
    }
    
    // Access custom fields
    if customField, ok := span.Raw["myField"]; ok {
        fmt.Println("Custom:", customField)
    }
}
```

### Error Handling

```go
doc, err := portabletext.Decode(reader)
if err != nil {
    // Check for specific error types
    var pErr *portabletext.Error
    if errors.As(err, &pErr) {
        fmt.Printf("Parse error at %s: %v\n", pErr.Path, pErr.Err)
        
        // Check underlying error
        if errors.Is(pErr.Err, portabletext.ErrMissingType) {
            fmt.Println("Missing _type field")
        }
    }
}
```

## Data Structures

### Node

Represents a Portable Text node (typically a block or custom object):

```go
type Node struct {
    Type     string              // Required: "_type"
    Key      string              // Optional: "_key"
    Style    *string             // Block style (e.g., "normal", "h1")
    Children []Span              // Child spans/inline objects
    MarkDefs []MarkDef           // Mark definitions (links, etc.)
    ListItem *string             // List item type
    Level    *int                // List nesting level
    Raw      map[string]any      // Unknown/custom fields
}
```

### Span

Represents an inline element within a block:

```go
type Span struct {
    Type  string              // Required: "_type"
    Text  *string             // Text content
    Marks []string            // Applied marks
    Raw   map[string]any      // Unknown/custom fields
}
```

### MarkDef

Represents a mark definition (annotations like links):

```go
type MarkDef struct {
    Key  string              // Required: "_key"
    Type string              // Required: "_type"
    Raw  map[string]any      // Unknown/custom fields (e.g., href)
}
```

## Examples

### Example 1: Extract All Text

```go
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
```

### Example 2: Find All Links

```go
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
```

### Example 3: Convert Headings

```go
func DowngradeHeadings(doc portabletext.Document) portabletext.Document {
    return portabletext.Transform(doc, func(n *portabletext.Node) *portabletext.Node {
        if !n.IsBlock() {
            return n
        }
        
        style := n.GetStyle()
        var newStyle string
        switch style {
        case "h1":
            newStyle = "h2"
        case "h2":
            newStyle = "h3"
        case "h3":
            newStyle = "h4"
        default:
            return n
        }
        
        n.Style = &newStyle
        return n
    })
}
```

### Example 4: Custom Validation

```go
func ValidateMaxLength(doc portabletext.Document, maxChars int) error {
    totalChars := 0
    for _, node := range doc {
        if node.IsBlock() {
            totalChars += len(node.GetText())
        }
    }
    
    if totalChars > maxChars {
        return fmt.Errorf("document exceeds %d characters: %d", maxChars, totalChars)
    }
    return nil
}
```

### Example 5: Add Table of Contents

```go
func GenerateTOC(doc portabletext.Document) []TOCEntry {
    var toc []TOCEntry
    
    portabletext.Walk(doc, func(n *portabletext.Node) error {
        style := n.GetStyle()
        if style == "h1" || style == "h2" || style == "h3" {
            toc = append(toc, TOCEntry{
                Level: style,
                Text:  n.GetText(),
                Key:   n.Key,
            })
        }
        return nil
    })
    
    return toc
}

type TOCEntry struct {
    Level string
    Text  string
    Key   string
}
```

## Portable Text Specification

This library implements the [Portable Text specification](https://github.com/portabletext/portabletext). Key concepts:

- **Block**: A top-level content node (paragraph, heading, etc.)
- **Span**: Inline text within a block with optional marks
- **Mark**: Text annotation (bold, italic, link, etc.)
- **MarkDef**: Definition of a mark with additional data (e.g., link URL)
- **Custom Objects**: Any node type beyond the standard spec

## Version History

### v0.1.1 (Current)

- Added `Key` field to `Node` for `_key` support
- Added builder functions: `NewBlock()`, `NewNode()`
- Added convenience methods: `GetStyle()`, `GetText()`, `GetListLevel()`, `AddSpan()`, `AddMarkDef()`
- Added `HasMark()` method to `Span`
- Added `DecodeString()` / `EncodeString()` convenience functions
- Added `WalkWithContext()` for contextual traversal
- Added `Filter()` and `Transform()` operations
- Enhanced validation with `ValidationOptions` and `ValidationError`
- Improved error types with structured `ValidationError`

### v0.1.0

- Initial release
- Core parsing and encoding
- Basic validation
- Error handling with paths
- Walk functionality
- Node cloning
- Raw field preservation

## Thread Safety

- **Reads**: Documents are safe for concurrent reads without synchronization
- **Writes**: Concurrent modifications require external synchronization
- **Immutability**: `Filter()` and `Transform()` create new documents and are safe to call concurrently

## Performance Considerations

- Decoding uses `json.NewDecoder` for efficient streaming
- All operations avoid unnecessary allocations where possible
- `Clone()` performs deep copies - use sparingly for large documents
- `Transform()` creates a new document - consider in-place modification for very large documents

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- [Sanity.io](https://www.sanity.io/) - Headless CMS using Portable Text
- [Portable Text](https://portabletext.org/) - Official specification

## Support

-  [Documentation](https://pkg.go.dev/github.com/derickschaefer/portabletext)
-  [Issue Tracker](https://github.com/derickschaefer/portabletext/issues)
-  [Discussions](https://github.com/derickschaefer/portabletext/discussions)
