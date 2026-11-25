package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	duckDuckGoLiteURL = "https://lite.duckduckgo.com/lite/"
	duckDuckGoAPIURL  = "https://api.duckduckgo.com/"
	searchHTTPClient  = &http.Client{Timeout: 15 * time.Second}
)

type searchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"`
}

func SearchWeb(query string) (string, error) {
	results, err := queryDDGLite(query)
	if err == nil && len(results) > 0 {
		return formatSearchResults(results), nil
	}

	answerResults, answerErr := queryDDGInstantAnswers(query)
	if answerErr == nil && len(answerResults) > 0 {
		return formatSearchResults(answerResults), nil
	}

	return "", fmt.Errorf("search failed (duckduckgo lite: %v, duckduckgo api: %v)", err, answerErr)
}

func formatSearchResults(results []searchResult) string {
	var out strings.Builder
	out.WriteString("Sources from DuckDuckGo:\n")
	for i, result := range results {
		fmt.Fprintf(&out, "[%d] %s\nURL: %s\nProvider: %s\nSnippet: %s\n\n", i+1, result.Title, result.URL, result.Source, result.Snippet)
	}
	return strings.TrimSpace(out.String())
}

func queryDDGLite(query string) ([]searchResult, error) {
	data := url.Values{}
	data.Set("q", query)
	data.Set("kl", "us-en")

	req, err := http.NewRequest("POST", duckDuckGoLiteURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://duckduckgo.com")
	req.Header.Set("Referer", "https://duckduckgo.com/")

	resp, err := searchHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	links := doc.Find("a.result-link")
	snippets := doc.Find("td.result-snippet")

	var results []searchResult
	links.Each(func(i int, s *goquery.Selection) {
		if len(results) >= 5 {
			return
		}
		title := strings.TrimSpace(s.Text())
		rawLink, _ := s.Attr("href")
		link := cleanDDGURL(rawLink)
		snippet := ""
		if i < snippets.Length() {
			snippet = cleanDDGSnippet(snippets.Eq(i).Text())
		}
		if title != "" && link != "" {
			results = append(results, searchResult{
				Title:   title,
				URL:     link,
				Snippet: snippet,
				Source:  "DuckDuckGo Lite",
			})
		}
	})

	return results, nil
}

func queryDDGInstantAnswers(query string) ([]searchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	req, err := http.NewRequest("GET", duckDuckGoAPIURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GoTerminalAgent/1.0 (web search tool)")

	resp, err := searchHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Heading       string `json:"Heading"`
		Answer        string `json:"Answer"`
		Abstract      string `json:"Abstract"`
		AbstractURL   string `json:"AbstractURL"`
		Definition    string `json:"Definition"`
		DefinitionURL string `json:"DefinitionURL"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var results []searchResult
	snippet := cleanDDGSnippet(payload.Answer)
	if snippet == "" {
		snippet = cleanDDGSnippet(payload.Abstract)
	}
	if snippet == "" {
		snippet = cleanDDGSnippet(payload.Definition)
	}
	link := strings.TrimSpace(payload.AbstractURL)
	if link == "" {
		link = strings.TrimSpace(payload.DefinitionURL)
	}
	if snippet != "" && link != "" {
		title := strings.TrimSpace(payload.Heading)
		if title == "" {
			title = query
		}
		results = append(results, searchResult{
			Title:   title,
			URL:     link,
			Snippet: snippet,
			Source:  "DuckDuckGo Instant Answers",
		})
	}

	for _, topic := range payload.RelatedTopics {
		if len(results) >= 5 {
			break
		}
		topicText := cleanDDGSnippet(topic.Text)
		topicURL := strings.TrimSpace(topic.FirstURL)
		if topicText != "" && topicURL != "" {
			results = append(results, searchResult{
				Title:   topicText,
				URL:     topicURL,
				Snippet: topicText,
				Source:  "DuckDuckGo Instant Answers",
			})
		}
	}

	return results, nil
}

func cleanDDGURL(raw string) string {
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if uddg := parsed.Query().Get("uddg"); uddg != "" {
		return uddg
	}
	return raw
}

func cleanDDGSnippet(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
	if err != nil {
		return strings.Join(strings.Fields(text), " ")
	}
	return strings.Join(strings.Fields(doc.Find("body").Text()), " ")
}

func ReadWebpage(targetURL string) (string, error) {
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// Strip non-content tags
	doc.Find("script, style, nav, footer, header, aside, noscript, iframe").Remove()

	// Collapse whitespace into clean paragraphs/text
	text := strings.Join(strings.Fields(doc.Find("body").Text()), " ")
	if text == "" {
		return fmt.Sprintf("Source: %s\nPage contained no readable text.", targetURL), nil
	}

	// Cap at ~6000 chars so it stays well within the 8192 token context window
	if len(text) > 6000 {
		text = text[:6000] + "... [content truncated]"
	}

	return fmt.Sprintf("Source: %s\n\n%s", targetURL, text), nil
}
