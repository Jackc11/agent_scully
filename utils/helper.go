package utils

import (
	"strings"

	"charm.land/glamour/v2"
)

func RenderMarkdown(md string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}
	return renderer.Render(md)
}

func ParseSlashCommand(input string) (string, string, bool) {
	if !strings.HasPrefix(input, "/") {
		return "", "", false
	}
	parts := strings.SplitN(strings.TrimSpace(input), " ", 2)
	cmd := strings.ToLower(parts[0])
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}
	return cmd, arg, true
}
