package main

import "testing"

func TestRegex_Match_Features(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    bool
	}{
		{"a", "a", true},
		{"a", "b", false},
		{"^hello", "hello world", true},
		{"^hello", "a hello", false},
		{"world$", "hello world", true},
		{"^hello$", "hello", true},
		{"^hello$", "hello world", false},
		{"h.llo", "hallo", true},
		{"h.llo", "hllo", false},
		{"[abc]+", "cab", true},
		{"[^abc]+", "def", true},
		{"a+b", "aaab", true},
		{"a+b", "b", false},
		{"ab?c", "abc", true},
		{"ab?c", "ac", true},
		{"ab?c", "abbc", false},
		{"(ab)+c", "ababc", true},
		{"(ab)+c", "aba", false},
		{"(ab)\\1", "abab", true},
		{"(ab)\\1", "abac", false},
		{"a|b", "b", true},
		{"a|b", "c", false},
		{"\\d+", "12345", true},
		{"\\w+", "abc_123", true},
		{"", "anything", true},
	}
	for _, tt := range tests {
		re, err := Compile(tt.pattern)
		if err != nil {
			t.Fatalf("Compile(%q) error: %v", tt.pattern, err)
		}
		got, err := re.Match([]byte(tt.text))
		if err != nil {
			t.Fatalf("Match(%q, %q) error: %v", tt.pattern, tt.text, err)
		}
		if got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.text, got, tt.want)
		}
	}
}

func TestRegex_Anchors_Exact(t *testing.T) {
	re, err := Compile("^foo$")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}
	if ok, _ := re.Match([]byte("foo")); !ok {
		t.Fatalf("expected exact match")
	}
	if ok, _ := re.Match([]byte("foobar")); ok {
		t.Fatalf("unexpected match for prefix")
	}
}

// Additional and more extensive tests for the regex engine
func TestRegex_ComplexPatterns(t *testing.T) {
	re, _ := Compile("^(foo|bar)+[abc]?$")
	cases := []struct {
		input string
		want  bool
	}{
		{"foo", true},
		{"bar", true},
		{"foobar", true},
		{"foofoo", true},
		{"barbarc", true},
		{"barbarx", false},
		{"", false},
	}
	for _, c := range cases {
		got, _ := re.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("input=%q: got %v, want %v", c.input, got, c.want)
		}
	}
}

func TestRegex_MultipleGroupsAndBackrefs(t *testing.T) {
	re, _ := Compile("(a)(b)\\2\\1")
	cases := []struct {
		input string
		want  bool
	}{
		{"abba", true},
		{"abab", false},
		{"ab", false},
	}
	for _, c := range cases {
		got, _ := re.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("input=%q: got %v, want %v", c.input, got, c.want)
		}
	}
}

func TestRegex_AlternationWithGroups(t *testing.T) {
	re, _ := Compile("(foo|bar)baz")
	cases := []struct {
		input string
		want  bool
	}{
		{"foobaz", true},
		{"barbaz", true},
		{"baz", false},
		{"foobar", false},
	}
	for _, c := range cases {
		got, _ := re.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("input=%q: got %v, want %v", c.input, got, c.want)
		}
	}
}

func TestRegex_CharClassEdgeCases(t *testing.T) {
	re, _ := Compile("[a]")
	cases := []struct {
		input string
		want  bool
	}{
		{"a", true},
		{"b", false},
		{"", false},
	}
	for _, c := range cases {
		got, _ := re.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("input=%q: got %v, want %v", c.input, got, c.want)
		}
	}

	reNeg, _ := Compile("[^a]")
	casesNeg := []struct {
		input string
		want  bool
	}{
		{"a", false},
		{"b", true},
		{"", false},
	}
	for _, c := range casesNeg {
		got, _ := reNeg.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("negated: input=%q: got %v, want %v", c.input, got, c.want)
		}
	}
}

func TestRegex_EmptyAndSpecialCases(t *testing.T) {
	re, _ := Compile("")
	cases := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"a", false},
		{" ", false},
	}
	for _, c := range cases {
		got, _ := re.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("input=%q: got %v, want %v", c.input, got, c.want)
		}
	}

	reDollar, _ := Compile("$")
	casesDollar := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"a", false},
	}
	for _, c := range casesDollar {
		got, _ := reDollar.Match([]byte(c.input))
		if got != c.want {
			t.Errorf("dollar: input=%q: got %v, want %v", c.input, got, c.want)
		}
	}
}

func TestRegexMatchFeatures(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    bool
	}{
		{"a", "a", true},
		{"a", "b", false},
		{"^hello", "hello world", true},
		{"^hello", "a hello", false},
		{"world$", "hello world", true},
		{"^hello$", "hello", true},
		{"^hello$", "hello world", false},
		{"h.llo", "hallo", true},
		{"h.llo", "hllo", false},
		{"[abc]+", "cab", true},
		{"[^abc]+", "def", true},
		{"a+b", "aaab", true},
		{"a+b", "b", false},
		{"ab?c", "abc", true},
		{"ab?c", "ac", true},
		{"ab?c", "abbc", false},
		{"(ab)+c", "ababc", true},
		{"(ab)+c", "aba", false},
		{"(ab)\\1", "abab", true},
		{"(ab)\\1", "abac", false},
		{"a|b", "b", true},
		{"a|b", "c", false},
		{"\\d+", "12345", true},
		{"\\w+", "abc_123", true},
		{"", "anything", true},
	}

	for _, tt := range tests {
		re, err := Compile(tt.pattern)
		if err != nil {
			t.Fatalf("Compile(%q) error: %v", tt.pattern, err)
		}
		got, err := re.Match([]byte(tt.text))
		if err != nil {
			t.Fatalf("Match(%q, %q) error: %v", tt.pattern, tt.text, err)
		}
		if got != tt.want {
			t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.text, got, tt.want)
		}
	}
}

func TestRegexAnchors(t *testing.T) {
	re, err := Compile("^foo$")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}
	if ok, _ := re.Match([]byte("foo")); !ok {
		t.Fatalf("expected exact match")
	}
	if ok, _ := re.Match([]byte("foobar")); ok {
		t.Fatalf("unexpected match for prefix")
	}
}
