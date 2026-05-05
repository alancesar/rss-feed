package markdown_test

import (
	"strings"
	"testing"

	"rss-feed/pkg/markdown"
	"rss-feed/pkg/rss"
)

func TestParser_Parse_DescriptionOnly(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Hello World",
		Description: "A short summary.",
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Hello World\nA short summary."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestParser_Parse_ContentPrependedBeforeDescription(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Hello World",
		Description: "A short summary.",
		Content:     "Full article body.",
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Hello World\nFull article body.\nA short summary."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestParser_Parse_AnchorInDescription_LabelPreserved(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Hello World",
		Description: `Check out <a href="https://example.com">this link</a> for more.`,
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Hello World\nCheck out this link for more."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestParser_Parse_AnchorInContent_LabelPreserved(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Hello World",
		Description: "A short summary.",
		Content:     `Read the <a href="https://example.com/full">full article</a> here.`,
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Hello World\nRead the full article here.\nA short summary."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestParser_Parse_MultipleAnchors_LabelsPreserved(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Links",
		Description: `Visit <a href="https://a.com">site A</a> and <a href="https://b.com">site B</a>.`,
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Links\nVisit site A and site B."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestParser_Parse_BlankLinesStripped(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Hello World",
		Description: "<p>First paragraph.</p><p>Second paragraph.</p>",
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, line := range strings.Split(got, "\n") {
		if strings.TrimSpace(line) == "" {
			t.Errorf("expected no blank lines in output, got %q", got)
		}
	}
}

func TestParser_Parse_HTMLFormattingConverted(t *testing.T) {
	parser := markdown.NewParser()
	article := rss.Article{
		Title:       "Formatting",
		Description: "This is <strong>bold</strong> and <em>italic</em>.",
	}

	got, err := parser.Parse(article)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Formatting\nThis is **bold** and *italic*."
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
