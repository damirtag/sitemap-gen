package parser

import (
	"bytes"
	"fmt"
	"net/url"

	"golang.org/x/net/html"
)

func ExtractLinks(body []byte, baseURL, allowedHost string) ([]string, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	var links []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					normalized, err := normalizeURL(base, a.Val, allowedHost)
					if err != nil {
						continue
					}
					if _, dup := seen[normalized]; !dup {
						seen[normalized] = struct{}{}
						links = append(links, normalized)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

// normalizeURL resolves href against base, strips fragment + tracking params,
// and returns empty string (error) if the result is off-domain.
func normalizeURL(base *url.URL, href, allowedHost string) (string, error) {
	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	resolved := base.ResolveReference(ref)
	resolved.Fragment = ""

	// only http/https
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", fmt.Errorf("skip scheme %s", resolved.Scheme)
	}

	// stay on the same domain
	if resolved.Hostname() != allowedHost {
		return "", fmt.Errorf("off-domain")
	}

	return resolved.String(), nil
}
