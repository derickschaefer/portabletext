package portabletext

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

//
// Public API
//

// Document is an ordered list of Portable Text nodes.
// Document is not safe for concurrent modification.
// For concurrent reads, no synchronization is needed.
// For concurrent writes, external synchronization is required.
type Document []Node

// Node represents a Portable Text node (block or custom object).
// Known fields are modeled; unknown/custom fields are preserved in Raw.
type Node struct {
	// Required
	Type string `json:"_type"`
	Key  string `json:"_key,omitempty"`

	// Common block fields
	Style    *string   `json:"style,omitempty"`
	Children []Span    `json:"children,omitempty"`
	MarkDefs []MarkDef `json:"markDefs,omitempty"`

	// List-related fields
	ListItem *string `json:"listItem,omitempty"`
	Level    *int    `json:"level,omitempty"`

	// Raw holds unknown/custom fields and preserves explicit nulls.
	Raw map[string]any `json:"-"`
}

// Span represents an inline node in a block's children array.
// Usually _type == "span", but inline objects are allowed too.
// For inline objects, Text is typically nil and Raw holds object fields.
type Span struct {
	Type  string   `json:"_type"`
	Text  *string  `json:"text,omitempty"`
	Marks []string `json:"marks,omitempty"`

	Raw map[string]any `json:"-"`
}

// MarkDef represents an annotation definition (e.g. link objects).
type MarkDef struct {
	Key  string `json:"_key"`
	Type string `json:"_type"`

	Raw map[string]any `json:"-"`
}

// ValidationOptions controls what Validate checks.
type ValidationOptions struct {
	RequireKeys      bool // Require _key on all blocks
	CheckMarkDefRefs bool // Verify mark references exist in markDefs
	AllowEmptyText   bool // Allow empty text in spans
}

// WalkContext provides context during tree traversal.
type WalkContext struct {
	Index      int
	Parent     *Node
	Depth      int
	BlockCount int
}

// Decode parses JSON Portable Text into a Document.
// - Requires _type on all nodes and child spans/markDefs where present
// - Captures unknown fields into Raw (including explicit nulls)
// - Does not normalize or semantically validate
func Decode(r io.Reader) (Document, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return nil, wrap("decode", "", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '[' {
		return nil, wrap("decode", "", fmt.Errorf("%w: expected '['", ErrUnexpectedToken))
	}

	var doc Document
	i := 0
	for dec.More() {
		var rm json.RawMessage
		if err := dec.Decode(&rm); err != nil {
			return nil, wrap("decode", fmt.Sprintf("[%d]", i), err)
		}
		n, err := parseNode(rm, fmt.Sprintf("[%d]", i))
		if err != nil {
			return nil, err
		}
		doc = append(doc, n)
		i++
	}

	tok, err = dec.Token()
	if err != nil {
		return nil, wrap("decode", "", err)
	}
	d, ok = tok.(json.Delim)
	if !ok || d != ']' {
		return nil, wrap("decode", "", fmt.Errorf("%w: expected ']'", ErrUnexpectedToken))
	}

	return doc, nil
}

// DecodeString is a convenience wrapper for Decode.
func DecodeString(s string) (Document, error) {
	return Decode(strings.NewReader(s))
}

// Encode serializes the AST back to JSON.
// - Re-emits all known and unknown fields
// - Does not mutate the input document
func Encode(w io.Writer, doc Document) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}

// EncodeString is a convenience wrapper for Encode.
func EncodeString(doc Document) (string, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Walk visits all top-level nodes in order; stops early on fn error.
func Walk(doc Document, fn func(*Node) error) error {
	for i := range doc {
		if err := fn(&doc[i]); err != nil {
			return err
		}
	}
	return nil
}

// WalkWithContext visits all top-level nodes with additional context.
func WalkWithContext(doc Document, fn func(*Node, WalkContext) error) error {
	blockCount := 0
	for i := range doc {
		ctx := WalkContext{
			Index:      i,
			Parent:     nil,
			Depth:      0,
			BlockCount: blockCount,
		}
		if doc[i].IsBlock() {
			blockCount++
		}
		if err := fn(&doc[i], ctx); err != nil {
			return err
		}
	}
	return nil
}

// Filter returns a new document with nodes matching the predicate.
func Filter(doc Document, pred func(*Node) bool) Document {
	result := make(Document, 0)
	for i := range doc {
		if pred(&doc[i]) {
			result = append(result, *doc[i].Clone())
		}
	}
	return result
}

// Transform applies fn to each node, returning a new document.
// If fn returns nil, the node is excluded from the result.
func Transform(doc Document, fn func(*Node) *Node) Document {
	result := make(Document, 0, len(doc))
	for i := range doc {
		if transformed := fn(doc[i].Clone()); transformed != nil {
			result = append(result, *transformed)
		}
	}
	return result
}

// Validate performs optional, opt-in checks. Unknown node types are never errors.
func Validate(doc Document) []error {
	return ValidateWithOptions(doc, ValidationOptions{})
}

// ValidateWithOptions performs validation with custom options.
func ValidateWithOptions(doc Document, opts ValidationOptions) []error {
	var errs []error
	for i := range doc {
		n := &doc[i]
		path := fmt.Sprintf("[%d]", i)

		if n.Type == "" {
			errs = append(errs, &ValidationError{
				Path:    path,
				Message: "missing _type",
				Node:    n,
			})
			continue
		}

		if opts.RequireKeys && n.Key == "" {
			errs = append(errs, &ValidationError{
				Path:    path,
				Message: "missing _key",
				Node:    n,
			})
		}

		if n.Type == "block" {
			// Build mark def map for reference checking
			var markDefMap map[string]bool
			if opts.CheckMarkDefRefs {
				markDefMap = make(map[string]bool)
				for _, md := range n.MarkDefs {
					markDefMap[md.Key] = true
				}
			}

			for j := range n.Children {
				c := &n.Children[j]
				cpath := fmt.Sprintf("%s.children[%d]", path, j)
				if c.Type == "" {
					errs = append(errs, &ValidationError{
						Path:    cpath,
						Message: "missing _type",
						Node:    n,
					})
					continue
				}
				if c.Type == "span" {
					if c.Text == nil {
						errs = append(errs, &ValidationError{
							Path:    cpath,
							Message: "span missing text",
							Node:    n,
						})
					} else if !opts.AllowEmptyText && *c.Text == "" {
						errs = append(errs, &ValidationError{
							Path:    cpath,
							Message: "span has empty text",
							Node:    n,
						})
					}

					// Check mark references
					if opts.CheckMarkDefRefs {
						for _, mark := range c.Marks {
							if !markDefMap[mark] {
								errs = append(errs, &ValidationError{
									Path:    cpath,
									Message: fmt.Sprintf("mark '%s' not found in markDefs", mark),
									Node:    n,
								})
							}
						}
					}
				}
			}

			for j := range n.MarkDefs {
				md := &n.MarkDefs[j]
				mdpath := fmt.Sprintf("%s.markDefs[%d]", path, j)
				if md.Type == "" {
					errs = append(errs, &ValidationError{
						Path:    mdpath,
						Message: "markDef missing _type",
						Node:    n,
					})
				}
				if md.Key == "" {
					errs = append(errs, &ValidationError{
						Path:    mdpath,
						Message: "markDef missing _key",
						Node:    n,
					})
				}
			}
		}
	}
	return errs
}

// IsBlock reports whether this node is a Portable Text "block".
func (n *Node) IsBlock() bool { return n != nil && n.Type == "block" }

// GetStyle returns the style or a default value.
func (n *Node) GetStyle() string {
	if n.Style != nil {
		return *n.Style
	}
	return "normal"
}

// GetText concatenates all span text in a block.
func (n *Node) GetText() string {
	var buf strings.Builder
	for _, child := range n.Children {
		if child.Text != nil {
			buf.WriteString(*child.Text)
		}
	}
	return buf.String()
}

// GetListLevel returns the list level or 1 if not set.
func (n *Node) GetListLevel() int {
	if n.Level != nil {
		return *n.Level
	}
	return 1
}

// Clone deep-copies the node, including Raw and nested slices/maps.
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}
	out := *n

	if n.Style != nil {
		s := *n.Style
		out.Style = &s
	}
	if n.ListItem != nil {
		s := *n.ListItem
		out.ListItem = &s
	}
	if n.Level != nil {
		l := *n.Level
		out.Level = &l
	}

	out.Children = cloneSpans(n.Children)
	out.MarkDefs = cloneMarkDefs(n.MarkDefs)
	out.Raw = deepCopyMap(n.Raw)

	return &out
}

// AddSpan adds a text span to a block node.
func (n *Node) AddSpan(text string, marks ...string) *Node {
	span := Span{
		Type:  "span",
		Text:  &text,
		Marks: marks,
		Raw:   map[string]any{},
	}
	n.Children = append(n.Children, span)
	return n
}

// AddMarkDef adds a mark definition to a block node.
func (n *Node) AddMarkDef(key, markType string, raw map[string]any) *Node {
	md := MarkDef{
		Key:  key,
		Type: markType,
		Raw:  raw,
	}
	if md.Raw == nil {
		md.Raw = map[string]any{}
	}
	n.MarkDefs = append(n.MarkDefs, md)
	return n
}

// HasMark checks if a span has a specific mark.
func (s *Span) HasMark(mark string) bool {
	for _, m := range s.Marks {
		if m == mark {
			return true
		}
	}
	return false
}

// NewBlock creates a basic block node.
func NewBlock(style string) *Node {
	return &Node{
		Type:     "block",
		Style:    &style,
		Children: []Span{},
		MarkDefs: []MarkDef{},
		Raw:      map[string]any{},
	}
}

// NewNode creates a custom node with the given type.
func NewNode(nodeType string) *Node {
	return &Node{
		Type: nodeType,
		Raw:  map[string]any{},
	}
}

//
// Errors (typed + path aware)
//

var (
	ErrMissingType     = errors.New("missing _type")
	ErrInvalidType     = errors.New("invalid _type")
	ErrExpectedObject  = errors.New("expected JSON object")
	ErrExpectedArray   = errors.New("expected JSON array")
	ErrInvalidMarks    = errors.New("marks must be an array of strings")
	ErrInvalidNumber   = errors.New("invalid number")
	ErrUnexpectedToken = errors.New("unexpected JSON token")
)

type Error struct {
	Op   string // "decode", "node", "span", "markDef"
	Path string // e.g. "[3].children[1].marks"
	Err  error
}

func (e *Error) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("portabletext %s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("portabletext %s at %s: %v", e.Op, e.Path, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

func wrap(op, path string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Op: op, Path: path, Err: err}
}

// ValidationError represents a validation error with context.
type ValidationError struct {
	Path    string
	Message string
	Node    *Node // Optional reference to problematic node
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

//
// Parsing (path aware)
//

func parseNode(b []byte, path string) (Node, error) {
	obj, err := decodeObjectUseNumber(b)
	if err != nil {
		return Node{}, wrap("node", path, err)
	}

	t, ok := obj["_type"]
	if !ok {
		return Node{}, wrap("node", path, ErrMissingType)
	}
	ts, ok := t.(string)
	if !ok || ts == "" {
		return Node{}, wrap("node", path, ErrInvalidType)
	}

	var n Node
	n.Type = ts
	n.Raw = map[string]any{}

	for k, v := range obj {
		switch k {
		case "_type":
		case "_key":
			if v == nil {
				n.Raw[k] = nil
				continue
			}
			if s, ok := v.(string); ok {
				n.Key = s
			} else {
				n.Raw[k] = v
			}
		case "style":
			if v == nil {
				n.Raw[k] = nil // preserve explicit null
				continue
			}
			if s, ok := v.(string); ok {
				n.Style = &s
			} else {
				n.Raw[k] = v
			}
		case "children":
			if v == nil {
				n.Raw[k] = nil // preserve explicit null
				continue
			}
			children, err := parseSpanArray(v, path+".children")
			if err != nil {
				return Node{}, err
			}
			n.Children = children
		case "markDefs":
			if v == nil {
				n.Raw[k] = nil // preserve explicit null
				continue
			}
			mds, err := parseMarkDefArray(v, path+".markDefs")
			if err != nil {
				return Node{}, err
			}
			n.MarkDefs = mds
		case "listItem":
			if v == nil {
				n.Raw[k] = nil
				continue
			}
			if s, ok := v.(string); ok {
				n.ListItem = &s
			} else {
				n.Raw[k] = v
			}
		case "level":
			if v == nil {
				n.Raw[k] = nil
				continue
			}
			switch x := v.(type) {
			case json.Number:
				iv, err := x.Int64()
				if err != nil {
					return Node{}, wrap("node", path+".level", ErrInvalidNumber)
				}
				i := int(iv)
				n.Level = &i
			default:
				n.Raw[k] = v
			}
		default:
			n.Raw[k] = v
		}
	}

	return n, nil
}

func parseSpanArray(v any, path string) ([]Span, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, wrap("node", path, ErrExpectedArray)
	}
	if len(arr) == 0 {
		return []Span{}, nil // preserve empty array
	}

	out := make([]Span, 0, len(arr))
	for i, item := range arr {
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, wrap("span", fmt.Sprintf("%s[%d]", path, i), err)
		}
		s, err := parseSpan(raw, fmt.Sprintf("%s[%d]", path, i))
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func parseSpan(b []byte, path string) (Span, error) {
	obj, err := decodeObjectUseNumber(b)
	if err != nil {
		return Span{}, wrap("span", path, err)
	}

	t, ok := obj["_type"]
	if !ok {
		return Span{}, wrap("span", path, ErrMissingType)
	}
	ts, ok := t.(string)
	if !ok || ts == "" {
		return Span{}, wrap("span", path, ErrInvalidType)
	}

	var s Span
	s.Type = ts
	s.Raw = map[string]any{}

	for k, v := range obj {
		switch k {
		case "_type":
		case "text":
			if v == nil {
				s.Raw[k] = nil // preserve explicit null
				continue
			}
			if str, ok := v.(string); ok {
				s.Text = &str
			} else {
				s.Raw[k] = v
			}
		case "marks":
			if v == nil {
				s.Raw[k] = nil // preserve explicit null
				continue
			}
			a, ok := v.([]any)
			if !ok {
				return Span{}, wrap("span", path+".marks", ErrInvalidMarks)
			}
			marks := make([]string, 0, len(a))
			for _, it := range a {
				ms, ok := it.(string)
				if !ok {
					return Span{}, wrap("span", path+".marks", ErrInvalidMarks)
				}
				marks = append(marks, ms)
			}
			s.Marks = marks // preserves empty array when present
		default:
			s.Raw[k] = v
		}
	}

	return s, nil
}

func parseMarkDefArray(v any, path string) ([]MarkDef, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, wrap("node", path, ErrExpectedArray)
	}
	if len(arr) == 0 {
		return []MarkDef{}, nil
	}

	out := make([]MarkDef, 0, len(arr))
	for i, item := range arr {
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, wrap("markDef", fmt.Sprintf("%s[%d]", path, i), err)
		}
		md, err := parseMarkDef(raw, fmt.Sprintf("%s[%d]", path, i))
		if err != nil {
			return nil, err
		}
		out = append(out, md)
	}
	return out, nil
}

func parseMarkDef(b []byte, path string) (MarkDef, error) {
	obj, err := decodeObjectUseNumber(b)
	if err != nil {
		return MarkDef{}, wrap("markDef", path, err)
	}

	t, ok := obj["_type"]
	if !ok {
		return MarkDef{}, wrap("markDef", path, ErrMissingType)
	}
	ts, ok := t.(string)
	if !ok || ts == "" {
		return MarkDef{}, wrap("markDef", path, ErrInvalidType)
	}

	var md MarkDef
	md.Type = ts
	md.Raw = map[string]any{}

	if k, ok := obj["_key"]; ok {
		if k == nil {
			md.Raw["_key"] = nil
		} else if ks, ok := k.(string); ok {
			md.Key = ks
		} else {
			md.Raw["_key"] = k
		}
	}

	for k, v := range obj {
		switch k {
		case "_type", "_key":
		default:
			md.Raw[k] = v
		}
	}

	return md, nil
}

func decodeObjectUseNumber(b []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	var obj map[string]any
	if err := dec.Decode(&obj); err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, ErrExpectedObject
	}
	return obj, nil
}

//
// JSON marshaling (re-emits Raw + known fields)
//

func (n Node) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(n.Raw)+8)

	for k, v := range n.Raw {
		m[k] = v
	}

	m["_type"] = n.Type

	if n.Key != "" {
		m["_key"] = n.Key
	}
	if n.Style != nil {
		m["style"] = *n.Style
	}
	if n.Children != nil {
		m["children"] = n.Children
	}
	if n.MarkDefs != nil {
		m["markDefs"] = n.MarkDefs
	}
	if n.ListItem != nil {
		m["listItem"] = *n.ListItem
	}
	if n.Level != nil {
		m["level"] = *n.Level
	}

	return json.Marshal(m)
}

func (s Span) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(s.Raw)+4)

	for k, v := range s.Raw {
		m[k] = v
	}

	m["_type"] = s.Type
	if s.Text != nil {
		m["text"] = *s.Text
	}
	if s.Marks != nil {
		m["marks"] = s.Marks
	}

	return json.Marshal(m)
}

func (md MarkDef) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(md.Raw)+3)

	for k, v := range md.Raw {
		m[k] = v
	}

	m["_type"] = md.Type
	if md.Key != "" {
		m["_key"] = md.Key
	}

	return json.Marshal(m)
}

//
// Deep copy helpers (for Clone)
//

func cloneSpans(in []Span) []Span {
	if in == nil {
		return nil
	}
	out := make([]Span, len(in))
	for i := range in {
		out[i] = in[i]
		if in[i].Text != nil {
			t := *in[i].Text
			out[i].Text = &t
		}
		if in[i].Marks != nil {
			out[i].Marks = append([]string(nil), in[i].Marks...)
		}
		out[i].Raw = deepCopyMap(in[i].Raw)
	}
	return out
}

func cloneMarkDefs(in []MarkDef) []MarkDef {
	if in == nil {
		return nil
	}
	out := make([]MarkDef, len(in))
	for i := range in {
		out[i] = in[i]
		out[i].Raw = deepCopyMap(in[i].Raw)
	}
	return out
}

func deepCopyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = deepCopyAny(v)
	}
	return out
}

func deepCopyAny(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return deepCopyMap(x)
	case []any:
		out := make([]any, len(x))
		for i := range x {
			out[i] = deepCopyAny(x[i])
		}
		return out
	case json.RawMessage:
		cp := make([]byte, len(x))
		copy(cp, x)
		return json.RawMessage(cp)
	case []byte:
		cp := make([]byte, len(x))
		copy(cp, x)
		return cp
	default:
		// primitives (string, bool, nil, json.Number, etc.)
		return x
	}
}
