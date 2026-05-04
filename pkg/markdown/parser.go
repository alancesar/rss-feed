package markdown

import (
	"fmt"
	"rss-feed/pkg/rss"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"golang.org/x/net/html"
)

type (
	Parser struct {
		conv *converter.Converter
	}
)

func NewParser() *Parser {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
	conv.Register.RendererFor("a", converter.TagTypeInline, func(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
		ctx.RenderChildNodes(ctx, w, n)
		return converter.RenderSuccess
	}, converter.PriorityEarly)

	return &Parser{
		conv: conv,
	}
}

func (p *Parser) Parse(article rss.Article) (string, error) {
	plain, err := p.conv.ConvertString(article.Description)
	if err != nil {
		return "", err
	}

	if article.Content != "" {
		content, err := p.conv.ConvertString(article.Content)
		if err != nil {
			return "", err
		}

		plain = content + "\n" + plain
	}

	plain = fmt.Sprintf("# %s\n%s", article.Title, plain)
	lines := strings.Split(plain, "\n")
	var result []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	output := strings.Join(result, "\n")
	return output, nil
}
