package sitemap

import (
	"encoding/xml"
	"os"
	"sync"
	"time"
)

type urlEntry struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

type urlSet struct {
	XMLName xml.Name   `xml:"urlset"`
	Xmlns   string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

// Map collects crawled URLs thread-safely, then writes sitemap.xml.
type Map struct {
	mu   sync.Mutex
	urls []urlEntry
}

func NewMap() *Map {
	return &Map{}
}

func (m *Map) Add(pageURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.urls = append(m.urls, urlEntry{
		Loc:        pageURL,
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "weekly",
		Priority:   "0.5",
	})
}

func (m *Map) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.urls)
}

func (m *Map) Write(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	set := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  m.urls,
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(xml.Header)
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	return enc.Encode(set)
}
