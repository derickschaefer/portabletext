## [0.1.2] - 2026-01-01

### Changed
- Updated README to reference working CLI examples in `/examples/` directory
- Improved documentation with links to actual example code
- Added code snippets for common operations

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-01-01

### Added

- `Key` field to `Node` struct for `_key` support
- Builder functions:
  - `NewBlock(style string)` - create a basic block
  - `NewNode(nodeType string)` - create a custom node
- Convenience methods on `Node`:
  - `GetStyle()` - returns style with "normal" default
  - `GetText()` - concatenates all span text
  - `GetListLevel()` - returns level with 1 as default
  - `AddSpan(text string, marks ...string)` - fluent API for adding spans
  - `AddMarkDef(key, markType string, raw map[string]any)` - fluent API for mark definitions
- `HasMark(mark string)` method on `Span` to check for specific marks
- String convenience functions:
  - `DecodeString(s string)` - decode from string
  - `EncodeString(doc Document)` - encode to string
- Traversal functions:
  - `WalkWithContext(doc Document, fn func(*Node, WalkContext) error)` - walk with context
  - `Filter(doc Document, pred func(*Node) bool)` - filter nodes by predicate
  - `Transform(doc Document, fn func(*Node) *Node)` - transform nodes to new document
- Validation improvements:
  - `ValidationOptions` struct for configurable validation
  - `ValidateWithOptions(doc Document, opts ValidationOptions)` - advanced validation
  - `ValidationError` struct with path, message, and node reference
  - Mark reference validation (checks marks exist in markDefs)
  - Optional `_key` requirement checking
  - Empty text validation
- Supporting types:
  - `WalkContext` - provides index, parent, depth, and block count during traversal
  - `ValidationError` - structured validation error with path and node

### Changed

- Enhanced validation error messages with structured `ValidationError` type
- Improved documentation with comprehensive examples
- Better error context in validation

### Fixed

- Parsing of `_key` field in Node (was not being captured in v0.1.0)

## [0.1.0] - 2026-01-01

### Added

- Initial release
- Core types:
  - `Document` - ordered list of nodes
  - `Node` - block or custom object
  - `Span` - inline text element
  - `MarkDef` - mark definition
- Decoding and encoding:
  - `Decode(r io.Reader)` - parse JSON to Document
  - `Encode(w io.Writer, doc Document)` - serialize Document to JSON
- Validation:
  - `Validate(doc Document)` - basic validation
  - Checks for required `_type` fields
  - Validates spans in blocks
  - Validates markDefs structure
- Traversal:
  - `Walk(doc Document, fn func(*Node) error)` - walk all nodes
- Node methods:
  - `IsBlock()` - check if node is a block
  - `Clone()` - deep copy of node
- Error handling:
  - `Error` type with operation and path context
  - Typed errors: `ErrMissingType`, `ErrInvalidType`, etc.
  - Path-aware error messages
- Raw field preservation:
  - Unknown/custom fields stored in `Raw` maps
  - Explicit null values preserved
  - Full round-trip fidelity
- Features:
  - Zero dependencies (standard library only)
  - Thread-safe for concurrent reads
  - Proper null handling with pointer fields
  - Preserves field order in JSON encoding

[0.1.1]: https://github.com/derickschaefer/portabletext/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/derickschaefer/portabletext/releases/tag/v0.1.0
