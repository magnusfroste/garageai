// Package kb is Bridge's knowledge base: docs/garage-troubleshooting.md, built into the binary so
// it matches this version's doctor and works when the garage cannot reach anything. One section per
// doctor problem code, plus general sections (updating, mix-ups).
package kb

import (
	_ "embed"
	"regexp"
	"sort"
	"strings"
)

//go:embed troubleshooting.md
var Text string

type Section struct {
	Code  string `json:"code,omitempty"` // the doctor problem code, "" for a general section
	Title string `json:"title"`
	Group string `json:"group"`
	Body  string `json:"body"`
}

var codeTitle = regexp.MustCompile("^`([a-z_]+)`")

// Sections splits the guide at its ## (group) and ### (section) headings. A ## heading with text
// before the first ### is a section of its own.
func Sections() []Section {
	var out []Section
	var group string
	var cur *Section
	flush := func() {
		if cur != nil {
			cur.Body = strings.TrimSpace(cur.Body)
			if cur.Body != "" {
				out = append(out, *cur)
			}
		}
		cur = nil
	}
	for _, line := range strings.Split(Text, "\n") {
		switch {
		case strings.HasPrefix(line, "### "):
			flush()
			title := strings.TrimPrefix(line, "### ")
			cur = &Section{Title: title, Group: group}
			if m := codeTitle.FindStringSubmatch(title); m != nil {
				cur.Code = m[1]
			}
		case strings.HasPrefix(line, "## "):
			flush()
			group = strings.TrimPrefix(line, "## ")
			cur = &Section{Title: group, Group: group}
		case strings.HasPrefix(line, "# "):
			// the document title
		default:
			if cur != nil {
				cur.Body += line + "\n"
			}
		}
	}
	flush()
	return out
}

// Explain returns the sections for a problem code: "heartbeat_missing" also finds the shared
// "`heartbeat_missing` / `heartbeat_stopped`" section.
func Explain(code string) []Section {
	var out []Section
	for _, s := range Sections() {
		if s.Code == code || strings.Contains(s.Title, "`"+code+"`") {
			out = append(out, s)
		}
	}
	return out
}

var word = regexp.MustCompile(`[a-z0-9_]{3,}`)

// Search ranks sections by how many of the query's words they contain (title words count double).
func Search(query string, limit int) []Section {
	terms := word.FindAllString(strings.ToLower(query), -1)
	type hit struct {
		s     Section
		score int
	}
	var hits []hit
	for _, s := range Sections() {
		title, body := strings.ToLower(s.Title), strings.ToLower(s.Body)
		score := 0
		for _, t := range terms {
			score += 2*strings.Count(title, t) + strings.Count(body, t)
		}
		if score > 0 {
			hits = append(hits, hit{s, score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	var out []Section
	for i := 0; i < len(hits) && i < limit; i++ {
		out = append(out, hits[i].s)
	}
	return out
}
