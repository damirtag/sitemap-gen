package report

import (
	"fmt"
	"strings"
	"time"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	white  = "\033[97m"
	gray   = "\033[90m"
)

type Stats struct {
	TargetURL   string
	OutputFile  string
	Total       int64
	Success     int64
	Errors      int64
	UniqueURLs  int
	SitemapURLs int
	Elapsed     time.Duration
	Workers     int
}

func PrintShutdown(s Stats) {
	width := 52
	bar := strings.Repeat("─", width)
	empty := strings.Repeat(" ", width)

	successRate := 0.0
	if s.Total > 0 {
		successRate = float64(s.Success) / float64(s.Total) * 100
	}

	urlsPerSec := 0.0
	if s.Elapsed.Seconds() > 0 {
		urlsPerSec = float64(s.Success) / s.Elapsed.Seconds()
	}

	fmt.Println()
	fmt.Printf("%s┌%s┐%s\n", cyan, bar, reset)
	fmt.Printf("%s│%s%s  sitemap-gen — crawl report%s%s  │%s\n", cyan, reset, bold+white, reset, strings.Repeat(" ", width-28), cyan)
	fmt.Printf("%s├%s┤%s\n", cyan, bar, reset)

	// target
	truncated := s.TargetURL
	if len(truncated) > width-14 {
		truncated = truncated[:width-17] + "..."
	}
	fmt.Printf("%s│%s  %starget%s   %s%-*s%s%s│%s\n",
		cyan, reset, gray, reset, white, width-12, truncated, reset, cyan, reset)

	// output file
	fmt.Printf("%s│%s  %soutput%s   %s%-*s%s%s│%s\n",
		cyan, reset, gray, reset, white, width-12, s.OutputFile, reset, cyan, reset)

	// elapsed + workers
	fmt.Printf("%s│%s  %selapsed%s  %s%-*s%s%s│%s\n",
		cyan, reset, gray, reset, white, width-12,
		fmt.Sprintf("%s  (%d workers)", s.Elapsed.Round(time.Millisecond), s.Workers),
		reset, cyan, reset)

	fmt.Printf("%s├%s┤%s\n", cyan, bar, reset)
	fmt.Printf("%s│%s%s  pages                                              │%s\n", cyan, reset, gray, reset)
	fmt.Printf("%s├%s┤%s\n", cyan, bar, reset)

	// success
	fmt.Printf("%s│%s  %s✓ crawled%s    %s%-*d%s%s│%s\n",
		cyan, reset, green, reset, white, width-14, s.Success, reset, cyan, reset)

	// errors
	errColor := green
	if s.Errors > 0 {
		errColor = red
	}
	fmt.Printf("%s│%s  %s✗ errors%s     %s%-*d%s%s│%s\n",
		cyan, reset, errColor, reset, white, width-14, s.Errors, reset, cyan, reset)

	// unique
	fmt.Printf("%s│%s  %s# unique%s     %s%-*d%s%s│%s\n",
		cyan, reset, yellow, reset, white, width-14, s.UniqueURLs, reset, cyan, reset)

	// sitemap entries
	fmt.Printf("%s│%s  %s↗ sitemap%s    %s%-*d%s%s│%s\n",
		cyan, reset, cyan, reset, white, width-14, s.SitemapURLs, reset, cyan, reset)

	fmt.Printf("%s├%s┤%s\n", cyan, bar, reset)

	// success rate bar
	filled := int(successRate / 100 * float64(width-4))
	unfilled := (width - 4) - filled
	rateColor := green
	if successRate < 50 {
		rateColor = red
	} else if successRate < 80 {
		rateColor = yellow
	}
	fmt.Printf("%s│%s  %s%s%s%s%s  │%s\n",
		cyan, reset,
		rateColor, strings.Repeat("█", filled),
		dim, strings.Repeat("░", unfilled),
		reset+cyan, reset)

	fmt.Printf("%s│%s  %ssuccess rate  %.1f%%%s%s%-*s%s%s│%s\n",
		cyan, reset, rateColor, successRate, reset,
		white, width-22, fmt.Sprintf("  %.1f urls/sec", urlsPerSec), reset, cyan, reset)

	fmt.Printf("%s└%s┘%s\n", cyan, bar, reset)

	// status line
	fmt.Println()
	if s.Errors == 0 {
		fmt.Printf("  %s%s✓ sitemap written to %s%s\n\n", bold+green, "", s.OutputFile, reset)
	} else {
		fmt.Printf("  %s%s⚠ sitemap written to %s (%d errors)%s\n\n", bold+yellow, "", s.OutputFile, s.Errors, reset)
	}

	_ = empty // used for padding reference
}
