package portabletext

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

// ========================================
// Decode Tests
// ========================================

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantLen int
	}{
		{
			name:    "basic block",
			input:   `[{"_type":"block","children":[{"_type":"span","text":"Hello"}],"markDefs":[]}]`,
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "empty document",
			input:   `[]`,
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "multiple blocks",
			input:   `[{"_type":"block","children":[],"markDefs":[]},{"_type":"block","children":[],"markDefs":[]}]`,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "invalid json",
			input:   `{not valid}`,
			wantErr: true,
		},
		{
			name:    "missing _type",
			input:   `[{"children":[]}]`,
			wantErr: true,
		},
		{
			name:    "not an array",
			input:   `{"_type":"block"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := DecodeString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(doc) != tt.wantLen {
				t.Errorf("DecodeString() len = %d, want %d", len(doc), tt.wantLen)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	input := `[{"_type":"block","_key":"key1","style":"h1","children":[{"_type":"span","text":"Title","marks":["strong"]}],"markDefs":[{"_type":"link","_key":"link1","href":"https://example.com"}],"listItem":"bullet","level":2}]`

	buf := bytes.NewBufferString(input)
	doc, err := Decode(buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if len(doc) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(doc))
	}

	node := doc[0]
	if node.Type != "block" {
		t.Errorf("Type = %s, want block", node.Type)
	}
	if node.Key != "key1" {
		t.Errorf("Key = %s, want key1", node.Key)
	}
	if node.GetStyle() != "h1" {
		t.Errorf("Style = %s, want h1", node.GetStyle())
	}
	if *node.ListItem != "bullet" {
		t.Errorf("ListItem = %s, want bullet", *node.ListItem)
	}
	if *node.Level != 2 {
		t.Errorf("Level = %d, want 2", *node.Level)
	}
	if len(node.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(node.Children))
	}
	if len(node.MarkDefs) != 1 {
		t.Fatalf("Expected 1 markDef, got %d", len(node.MarkDefs))
	}
}

func TestDecodeWithNulls(t *testing.T) {
	input := `[{"_type":"block","style":null,"children":null,"markDefs":null}]`

	doc, err := DecodeString(input)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	node := doc[0]
	if _, ok := node.Raw["style"]; !ok {
		t.Error("Expected explicit null in Raw for style")
	}
}

func TestDecodeCustomFields(t *testing.T) {
	input := `[{"_type":"block","customField":"value","customNumber":42,"customBool":true,"children":[],"markDefs":[]}]`

	doc, err := DecodeString(input)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	node := doc[0]
	if node.Raw["customField"] != "value" {
		t.Error("Custom string field not preserved")
	}
	if node.Raw["customBool"] != true {
		t.Error("Custom bool field not preserved")
	}
}

func TestDecodeInvalidLevel(t *testing.T) {
	input := `[{"_type":"block","level":"not a number","children":[],"markDefs":[]}]`

	doc, err := DecodeString(input)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	// Should preserve in Raw when not a valid number
	if _, ok := doc[0].Raw["level"]; !ok {
		t.Error("Invalid level should be in Raw")
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "empty type",
			input:     `[{"_type":""}]`,
			wantError: "invalid _type",
		},
		{
			name:      "null type",
			input:     `[{"_type":null}]`,
			wantError: "invalid _type",
		},
		{
			name:      "missing type in span",
			input:     `[{"_type":"block","children":[{"text":"hi"}],"markDefs":[]}]`,
			wantError: "missing _type",
		},
		{
			name:      "invalid marks",
			input:     `[{"_type":"block","children":[{"_type":"span","text":"hi","marks":"not-array"}],"markDefs":[]}]`,
			wantError: "marks must be an array",
		},
		{
			name:      "marks with non-string",
			input:     `[{"_type":"block","children":[{"_type":"span","text":"hi","marks":[123]}],"markDefs":[]}]`,
			wantError: "marks must be an array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeString(tt.input)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}

// ========================================
// Encode Tests
// ========================================

func TestEncodeString(t *testing.T) {
	doc := Document{*NewBlock("normal").AddSpan("Hello")}
	result, err := EncodeString(doc)
	if err != nil {
		t.Fatalf("EncodeString failed: %v", err)
	}
	if result == "" {
		t.Error("EncodeString returned empty string")
	}

	// Verify it's valid JSON
	var decoded []map[string]any
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Errorf("EncodeString produced invalid JSON: %v", err)
	}
}

func TestEncode(t *testing.T) {
	doc := Document{*NewBlock("h1").AddSpan("Title", "strong")}

	var buf bytes.Buffer
	err := Encode(&buf, doc)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Encode() produced empty output")
	}
}

func TestRoundTrip(t *testing.T) {
	input := `[{"_type":"block","_key":"abc123","style":"h1","children":[{"_type":"span","text":"Hello","marks":["strong"]}],"markDefs":[{"_type":"link","_key":"link1","href":"https://example.com"}]}]`

	doc, err := DecodeString(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	output, err := EncodeString(doc)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode again to verify
	doc2, err := DecodeString(output)
	if err != nil {
		t.Fatalf("Second decode failed: %v", err)
	}

	if len(doc2) != len(doc) {
		t.Errorf("Round trip changed document length: %d -> %d", len(doc), len(doc2))
	}

	// Check that custom fields survived
	if href, ok := doc2[0].MarkDefs[0].Raw["href"].(string); !ok || href != "https://example.com" {
		t.Error("Round trip did not preserve custom markDef fields")
	}
}

func TestEncodePreservesRaw(t *testing.T) {
	node := NewBlock("normal")
	node.Raw["custom"] = "value"
	node.Raw["number"] = json.Number("42")

	doc := Document{*node}
	output, err := EncodeString(doc)
	if err != nil {
		t.Fatalf("EncodeString() error = %v", err)
	}

	// Decode and verify
	doc2, err := DecodeString(output)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	if doc2[0].Raw["custom"] != "value" {
		t.Error("Custom field not preserved")
	}
}

// ========================================
// Validation Tests
// ========================================

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		doc      Document
		wantErrs int
	}{
		{
			name:     "valid block",
			doc:      Document{*NewBlock("normal").AddSpan("Hello")},
			wantErrs: 0,
		},
		{
			name: "missing _type",
			doc: Document{Node{
				Type:     "",
				Children: []Span{},
			}},
			wantErrs: 1,
		},
		{
			name: "span missing text",
			doc: Document{Node{
				Type: "block",
				Children: []Span{
					{Type: "span", Text: nil},
				},
			}},
			wantErrs: 1,
		},
		{
			name: "span missing _type",
			doc: Document{Node{
				Type: "block",
				Children: []Span{
					{Type: ""},
				},
			}},
			wantErrs: 1,
		},
		{
			name: "markDef missing _type",
			doc: Document{Node{
				Type:     "block",
				Children: []Span{},
				MarkDefs: []MarkDef{
					{Type: "", Key: "test"},
				},
			}},
			wantErrs: 1,
		},
		{
			name: "markDef missing _key",
			doc: Document{Node{
				Type:     "block",
				Children: []Span{},
				MarkDefs: []MarkDef{
					{Type: "link", Key: ""},
				},
			}},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(tt.doc)
			if len(errs) != tt.wantErrs {
				t.Errorf("Validate() got %d errors, want %d", len(errs), tt.wantErrs)
				for _, err := range errs {
					t.Logf("  Error: %v", err)
				}
			}
		})
	}
}

func TestValidateWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		doc      Document
		opts     ValidationOptions
		wantErrs int
	}{
		{
			name:     "require keys - missing",
			doc:      Document{*NewBlock("normal").AddSpan("Hello")},
			opts:     ValidationOptions{RequireKeys: true},
			wantErrs: 1,
		},
		{
			name: "require keys - present",
			doc: Document{Node{
				Type:     "block",
				Key:      "key1",
				Children: []Span{{Type: "span", Text: stringPtr("Hi")}},
			}},
			opts:     ValidationOptions{RequireKeys: true},
			wantErrs: 0,
		},
		{
			name: "check mark refs - invalid",
			doc: Document{Node{
				Type: "block",
				Children: []Span{
					{Type: "span", Text: stringPtr("Hi"), Marks: []string{"nonexistent"}},
				},
				MarkDefs: []MarkDef{},
			}},
			opts:     ValidationOptions{CheckMarkDefRefs: true},
			wantErrs: 1,
		},
		{
			name: "check mark refs - valid",
			doc: Document{Node{
				Type: "block",
				Children: []Span{
					{Type: "span", Text: stringPtr("Hi"), Marks: []string{"link1"}},
				},
				MarkDefs: []MarkDef{
					{Type: "link", Key: "link1"},
				},
			}},
			opts:     ValidationOptions{CheckMarkDefRefs: true},
			wantErrs: 0,
		},
		{
			name: "allow empty text - false",
			doc: Document{Node{
				Type: "block",
				Children: []Span{
					{Type: "span", Text: stringPtr("")},
				},
			}},
			opts:     ValidationOptions{AllowEmptyText: false},
			wantErrs: 1,
		},
		{
			name: "allow empty text - true",
			doc: Document{Node{
				Type: "block",
				Children: []Span{
					{Type: "span", Text: stringPtr("")},
				},
			}},
			opts:     ValidationOptions{AllowEmptyText: true},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateWithOptions(tt.doc, tt.opts)
			if len(errs) != tt.wantErrs {
				t.Errorf("ValidateWithOptions() got %d errors, want %d", len(errs), tt.wantErrs)
				for _, err := range errs {
					t.Logf("  Error: %v", err)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	ve := &ValidationError{
		Path:    "[0].children[1]",
		Message: "span missing text",
	}

	expected := "[0].children[1]: span missing text"
	if ve.Error() != expected {
		t.Errorf("ValidationError.Error() = %s, want %s", ve.Error(), expected)
	}
}

// ========================================
// Node Methods Tests
// ========================================

func TestNodeMethods(t *testing.T) {
	node := NewBlock("h1").AddSpan("Hello")

	if !node.IsBlock() {
		t.Error("Node.IsBlock() should be true")
	}

	if node.GetStyle() != "h1" {
		t.Errorf("GetStyle() = %s, want h1", node.GetStyle())
	}

	if node.GetText() != "Hello" {
		t.Errorf("GetText() = %s, want Hello", node.GetText())
	}

	if node.GetListLevel() != 1 {
		t.Errorf("GetListLevel() = %d, want 1", node.GetListLevel())
	}
}

func TestNodeGetStyleDefault(t *testing.T) {
	node := &Node{Type: "block"}
	if node.GetStyle() != "normal" {
		t.Errorf("GetStyle() with nil style = %s, want normal", node.GetStyle())
	}
}

func TestNodeGetListLevelWithValue(t *testing.T) {
	level := 3
	node := &Node{Type: "block", Level: &level}
	if node.GetListLevel() != 3 {
		t.Errorf("GetListLevel() = %d, want 3", node.GetListLevel())
	}
}

func TestNodeGetTextMultipleSpans(t *testing.T) {
	node := NewBlock("normal").
		AddSpan("Hello ").
		AddSpan("world").
		AddSpan("!")

	expected := "Hello world!"
	if node.GetText() != expected {
		t.Errorf("GetText() = %s, want %s", node.GetText(), expected)
	}
}

func TestNodeGetTextWithNilText(t *testing.T) {
	node := &Node{
		Type: "block",
		Children: []Span{
			{Type: "span", Text: stringPtr("Hello")},
			{Type: "span", Text: nil}, // Inline object
			{Type: "span", Text: stringPtr("World")},
		},
	}

	expected := "HelloWorld"
	if node.GetText() != expected {
		t.Errorf("GetText() = %s, want %s", node.GetText(), expected)
	}
}

func TestNodeIsBlockNil(t *testing.T) {
	var node *Node
	if node.IsBlock() {
		t.Error("IsBlock() on nil node should be false")
	}
}

func TestNodeAddSpan(t *testing.T) {
	node := NewBlock("normal")
	node.AddSpan("Hello", "strong", "em")

	if len(node.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(node.Children))
	}

	span := node.Children[0]
	if *span.Text != "Hello" {
		t.Errorf("Text = %s, want Hello", *span.Text)
	}
	if len(span.Marks) != 2 {
		t.Errorf("Expected 2 marks, got %d", len(span.Marks))
	}
}

func TestNodeAddMarkDef(t *testing.T) {
	node := NewBlock("normal")
	node.AddMarkDef("link1", "link", map[string]any{"href": "https://example.com"})

	if len(node.MarkDefs) != 1 {
		t.Fatalf("Expected 1 markDef, got %d", len(node.MarkDefs))
	}

	md := node.MarkDefs[0]
	if md.Key != "link1" {
		t.Errorf("Key = %s, want link1", md.Key)
	}
	if md.Type != "link" {
		t.Errorf("Type = %s, want link", md.Type)
	}
	if md.Raw["href"] != "https://example.com" {
		t.Error("href not preserved in Raw")
	}
}

func TestNodeAddMarkDefNilRaw(t *testing.T) {
	node := NewBlock("normal")
	node.AddMarkDef("key1", "type1", nil)

	if node.MarkDefs[0].Raw == nil {
		t.Error("AddMarkDef should initialize Raw if nil")
	}
}

func TestNodeClone(t *testing.T) {
	original := NewBlock("normal").AddSpan("Hello", "strong")
	original.Key = "key1"
	original.Raw["custom"] = "value"

	clone := original.Clone()

	// Modify clone
	clone.AddSpan(" World")
	clone.Key = "key2"
	clone.Raw["custom"] = "changed"

	// Original should be unchanged
	if len(original.Children) == len(clone.Children) {
		t.Error("Clone was not independent - children affected")
	}
	if original.Key == clone.Key {
		t.Error("Clone was not independent - key affected")
	}
	if original.Raw["custom"] == clone.Raw["custom"] {
		t.Error("Clone was not independent - Raw affected")
	}
}

func TestNodeCloneNil(t *testing.T) {
	var node *Node
	clone := node.Clone()
	if clone != nil {
		t.Error("Cloning nil node should return nil")
	}
}

func TestNodeClonePointerFields(t *testing.T) {
	style := "h1"
	listItem := "bullet"
	level := 2

	original := &Node{
		Type:     "block",
		Style:    &style,
		ListItem: &listItem,
		Level:    &level,
	}

	clone := original.Clone()

	// Modify original's pointer values
	*original.Style = "h2"
	*original.ListItem = "number"
	*original.Level = 3

	// Clone should be unchanged
	if *clone.Style != "h1" {
		t.Error("Clone style was affected by original modification")
	}
	if *clone.ListItem != "bullet" {
		t.Error("Clone listItem was affected by original modification")
	}
	if *clone.Level != 2 {
		t.Error("Clone level was affected by original modification")
	}
}

// ========================================
// Span Tests
// ========================================

func TestSpanHasMark(t *testing.T) {
	span := Span{
		Type:  "span",
		Marks: []string{"strong", "em"},
	}

	if !span.HasMark("strong") {
		t.Error("HasMark(strong) should be true")
	}

	if !span.HasMark("em") {
		t.Error("HasMark(em) should be true")
	}

	if span.HasMark("underline") {
		t.Error("HasMark(underline) should be false")
	}
}

func TestSpanHasMarkEmptyMarks(t *testing.T) {
	span := Span{Type: "span", Marks: []string{}}
	if span.HasMark("strong") {
		t.Error("HasMark on empty marks should be false")
	}
}

func TestSpanHasMarkNilMarks(t *testing.T) {
	span := Span{Type: "span", Marks: nil}
	if span.HasMark("strong") {
		t.Error("HasMark on nil marks should be false")
	}
}

// ========================================
// Builder Tests
// ========================================

func TestNewBlock(t *testing.T) {
	block := NewBlock("h1")

	if block.Type != "block" {
		t.Errorf("Type = %s, want block", block.Type)
	}
	if block.GetStyle() != "h1" {
		t.Errorf("Style = %s, want h1", block.GetStyle())
	}
	if block.Children == nil {
		t.Error("Children should be initialized")
	}
	if block.MarkDefs == nil {
		t.Error("MarkDefs should be initialized")
	}
	if block.Raw == nil {
		t.Error("Raw should be initialized")
	}
}

func TestNewNode(t *testing.T) {
	node := NewNode("customType")

	if node.Type != "customType" {
		t.Errorf("Type = %s, want customType", node.Type)
	}
	if node.Raw == nil {
		t.Error("Raw should be initialized")
	}
}

// ========================================
// Walk Tests
// ========================================

func TestWalk(t *testing.T) {
	doc := Document{
		*NewBlock("h1").AddSpan("Title"),
		*NewBlock("normal").AddSpan("Content"),
		*NewNode("customType"),
	}

	count := 0
	err := Walk(doc, func(node *Node) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("Walk() returned error: %v", err)
	}

	if count != 3 {
		t.Errorf("Walk() visited %d nodes, want 3", count)
	}
}

func TestWalkEarlyStop(t *testing.T) {
	doc := Document{
		*NewBlock("h1"),
		*NewBlock("h2"),
		*NewBlock("h3"),
	}

	count := 0
	testErr := errors.New("stop")
	err := Walk(doc, func(node *Node) error {
		count++
		if count == 2 {
			return testErr
		}
		return nil
	})

	if err != testErr {
		t.Errorf("Walk() error = %v, want %v", err, testErr)
	}

	if count != 2 {
		t.Errorf("Walk() should stop at 2, got %d", count)
	}
}

func TestWalkWithContext(t *testing.T) {
	doc := Document{
		*NewBlock("h1").AddSpan("Title"),
		*NewNode("customType"),
		*NewBlock("normal").AddSpan("Content"),
	}

	indices := []int{}
	blockCounts := []int{}

	err := WalkWithContext(doc, func(node *Node, ctx WalkContext) error {
		indices = append(indices, ctx.Index)
		blockCounts = append(blockCounts, ctx.BlockCount)

		if ctx.Index < 0 || ctx.Index >= len(doc) {
			t.Errorf("Invalid index in context: %d", ctx.Index)
		}
		if ctx.Parent != nil {
			t.Error("Top-level walk should have nil parent")
		}
		if ctx.Depth != 0 {
			t.Error("Top-level walk should have depth 0")
		}
		return nil
	})

	if err != nil {
		t.Errorf("WalkWithContext() returned error: %v", err)
	}

	if len(indices) != 3 {
		t.Errorf("Expected 3 indices, got %d", len(indices))
	}

	// Check block counts increase only for blocks
	if blockCounts[0] != 0 || blockCounts[1] != 1 || blockCounts[2] != 1 {
		t.Errorf("Block counts incorrect: %v", blockCounts)
	}
}

// ========================================
// Filter and Transform Tests
// ========================================

func TestFilter(t *testing.T) {
	doc := Document{
		*NewBlock("h1").AddSpan("Title"),
		*NewNode("customType"),
		*NewBlock("normal").AddSpan("Content"),
		*NewNode("anotherCustom"),
	}

	blocks := Filter(doc, func(n *Node) bool {
		return n.IsBlock()
	})

	if len(blocks) != 2 {
		t.Errorf("Filter() returned %d blocks, want 2", len(blocks))
	}

	for _, node := range blocks {
		if !node.IsBlock() {
			t.Error("Filter() returned non-block")
		}
	}
}

func TestFilterClones(t *testing.T) {
	original := NewBlock("h1")
	original.Raw["custom"] = "value"
	doc := Document{*original}

	filtered := Filter(doc, func(n *Node) bool { return true })

	// Modify filtered
	filtered[0].Raw["custom"] = "changed"

	// Original should be unchanged
	if original.Raw["custom"] == "changed" {
		t.Error("Filter() did not clone nodes")
	}
}

func TestTransform(t *testing.T) {
	doc := Document{
		*NewBlock("h1").AddSpan("Title"),
		*NewBlock("h2").AddSpan("Subtitle"),
		*NewBlock("normal").AddSpan("Content"),
	}

	transformed := Transform(doc, func(n *Node) *Node {
		style := n.GetStyle()
		if style == "h1" || style == "h2" {
			newStyle := "h3"
			n.Style = &newStyle
		}
		return n
	})

	if len(transformed) != 3 {
		t.Errorf("Transform() returned %d nodes, want 3", len(transformed))
	}

	if transformed[0].GetStyle() != "h3" {
		t.Error("Transform() did not change h1 to h3")
	}
	if transformed[1].GetStyle() != "h3" {
		t.Error("Transform() did not change h2 to h3")
	}
	if transformed[2].GetStyle() != "normal" {
		t.Error("Transform() should not change normal")
	}
}

func TestTransformExcludeNil(t *testing.T) {
	doc := Document{
		*NewBlock("h1"),
		*NewBlock("h2"),
		*NewBlock("normal"),
	}

	transformed := Transform(doc, func(n *Node) *Node {
		if n.GetStyle() == "h2" {
			return nil // Exclude h2
		}
		return n
	})

	if len(transformed) != 2 {
		t.Errorf("Transform() returned %d nodes, want 2 (h2 excluded)", len(transformed))
	}
}

// ========================================
// Error Tests
// ========================================

func TestErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &Error{
		Op:   "decode",
		Path: "[0]",
		Err:  innerErr,
	}

	if !errors.Is(err, innerErr) {
		t.Error("Error.Unwrap() not working with errors.Is()")
	}
}

func TestErrorString(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "with path",
			err: &Error{
				Op:   "decode",
				Path: "[0].children[1]",
				Err:  ErrMissingType,
			},
			expected: "portabletext decode at [0].children[1]: missing _type",
		},
		{
			name: "without path",
			err: &Error{
				Op:  "encode",
				Err: errors.New("test error"),
			},
			expected: "portabletext encode: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error.Error() = %s, want %s", tt.err.Error(), tt.expected)
			}
		})
	}
}

// ========================================
// Helper Functions
// ========================================

func stringPtr(s string) *string {
	return &s
}
