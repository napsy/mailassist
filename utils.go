package main

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func highlightPriority(msg string) string {
	msg = strings.ReplaceAll(msg, "low", `<span style="font-weight: bold; background: black; color: #99ce88;">low</span>`)
	msg = strings.ReplaceAll(msg, "medium", `<span style="font-weight: bold; background: black; color: #49a8fc;">medium</span>`)
	msg = strings.ReplaceAll(msg, "med", `<span style="font-weight: bold; background: black; color: #49a8fc;">med</span>`)
	msg = strings.ReplaceAll(msg, "high", `<span style="font-weight: bold; background: black; color: #fc6764;">high</span>`)

	msg = strings.ReplaceAll(msg, "Low", `<span style="font-weight: bold; background: black; color: #99ce88;">Low</span>`)
	msg = strings.ReplaceAll(msg, "Medium", `<span style="font-weight: bold; background: black; color: #49a8fc;">Medium</span>`)
	msg = strings.ReplaceAll(msg, "Med", `<span style="font-weight: bold; background: black; color: #49a8fc;">Med</span>`)
	msg = strings.ReplaceAll(msg, "High", `<span style="font-weight: bold; background: black; color: #fc6764;">High</span>`)

	return msg
}

func markdownMessage(msg string) string {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(msg))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}
