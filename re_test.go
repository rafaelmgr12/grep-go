package main

import "testing"

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
