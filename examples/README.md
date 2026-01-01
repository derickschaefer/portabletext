# Portable Text Examples

This directory contains practical examples of using the portabletext package.

## Sample Data Files

- **basic.json** - Simple document with headings, paragraphs, lists, and links
- **blog-post.json** - Blog post with custom nodes (image, code blocks)

## CLI Examples

Each example is a standalone program demonstrating different use cases.

### 1. Extract Text (`01-extract-text`)

Extracts plain text from Portable Text JSON with simple formatting.

```bash
cd 01-extract-text
go run main.go ../basic.json
```

**Output:**
```
# Getting Started with Portable Text

Portable Text is a JSON-based rich text specification that represents content as an abstract syntax tree.

## Key Features

  • Structured content representation
  • Support for annotations and marks
  • Extensible with custom objects

For more information, visit the official documentation.
```

### 2. Find Links (`02-find-links`)

Extracts all links with their associated text and metadata.

```bash
cd 02-find-links
go run main.go ../basic.json
```

**Output:**
```
[1] Key: link1
    URL: https://portabletext.org
    Text: official documentation

Total links: 1
```

### 3. Build Document (`03-build-document`)

Demonstrates programmatic document creation.

```bash
cd 03-build-document
go run main.go > output.json
```

Creates a complete Portable Text document from scratch using the builder API.

### 4. Transform Headings (`04-transform-headings`)

Upgrades or downgrades heading levels.

```bash
cd 04-transform-headings

# Downgrade all headings (h1 -> h2, h2 -> h3, etc)
go run main.go --downgrade ../basic.json

# Upgrade all headings (h2 -> h1, h3 -> h2, etc)
go run main.go ../basic.json
```

### 5. Table of Contents (`05-table-of-contents`)

Generates a table of contents from document headings.

```bash
cd 05-table-of-contents

# Plain text format
go run main.go ../basic.json

# Markdown format
go run main.go --markdown ../basic.json
```

**Output (plain):**
```
Table of Contents:
==================================================
1. Getting Started with Portable Text [block1]
  2. Key Features [block3]
```

**Output (markdown):**
```
## Table of Contents

- [Getting Started with Portable Text](#getting-started-with-portable-text)
  - [Key Features](#key-features)
```

## Running Examples

```bash
# Install the package first
go get github.com/derickschaefer/portabletext

# Run any example
cd examples/01-extract-text
go run main.go ../basic.json
```

## Common Patterns

### Reading a File

```go
file, err := os.Open(filename)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

doc, err := portabletext.Decode(file)
```

### Walking Nodes

```go
portabletext.Walk(doc, func(node *portabletext.Node) error {
    // Process each node
    return nil
})
```

### Filtering

```go
blocks := portabletext.Filter(doc, func(n *portabletext.Node) bool {
    return n.IsBlock()
})
```

### Transforming

```go
transformed := portabletext.Transform(doc, func(n *portabletext.Node) *portabletext.Node {
    // Modify node
    return n
})
```

## Use Cases

These examples demonstrate common use cases:

1. **Content Migration** - Extract text for migration to other formats
2. **SEO Analysis** - Extract links and metadata
3. **CMS Integration** - Build content programmatically
4. **Content Transformation** - Modify structure (headings, styles)
5. **Navigation** - Generate TOC for documentation sites

## Contributing

Feel free to add more examples! Each example should:
- Be self-contained in its own directory
- Include clear documentation
- Demonstrate a practical use case
- Use the package idiomatically
