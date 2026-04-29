package parser

import (
	"testing"
)

func TestExtractLinks_AbsoluteURL(t *testing.T) {
	body := []byte(`<a href="https://example.com/about">about</a>`)
	links, err := ExtractLinks(body, "https://example.com", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != "https://example.com/about" {
		t.Errorf("unexpected links: %v", links)
	}
}

func TestExtractLinks_RelativeURL(t *testing.T) {
	body := []byte(`<a href="/about">about</a>`)
	links, err := ExtractLinks(body, "https://example.com/page", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != "https://example.com/about" {
		t.Errorf("unexpected links: %v", links)
	}
}

func TestExtractLinks_OffDomainFiltered(t *testing.T) {
	body := []byte(`<a href="https://other.com/page">other</a>`)
	links, err := ExtractLinks(body, "https://example.com", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Errorf("expected off-domain link to be filtered, got: %v", links)
	}
}

func TestExtractLinks_FragmentStripped(t *testing.T) {
	body := []byte(`<a href="/page#section">link</a>`)
	links, err := ExtractLinks(body, "https://example.com", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0] != "https://example.com/page" {
		t.Errorf("fragment not stripped: %v", links)
	}
}

func TestExtractLinks_Deduplication(t *testing.T) {
	body := []byte(`
        <a href="/page">one</a>
        <a href="/page">two</a>
    `)
	links, err := ExtractLinks(body, "https://example.com", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Errorf("expected 1 deduplicated link, got %d", len(links))
	}
}

func TestExtractLinks_SkipsNonHTTP(t *testing.T) {
	body := []byte(`
        <a href="mailto:hi@example.com">mail</a>
        <a href="tel:+1234">phone</a>
        <a href="/valid">valid</a>
    `)
	links, err := ExtractLinks(body, "https://example.com", "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Errorf("expected only 1 valid link, got %d", len(links))
	}
}
