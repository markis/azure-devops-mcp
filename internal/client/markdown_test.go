package client

import "testing"

func TestConvertMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"empty string",
			"",
			"",
		},
		{
			"plain text",
			"hello world",
			"<p>hello world</p>\n",
		},
		{
			"bold text",
			"**bold** text",
			"<p><strong>bold</strong> text</p>\n",
		},
		{
			"italic text",
			"*italic* text",
			"<p><em>italic</em> text</p>\n",
		},
		{
			"header",
			"## Header\n\nParagraph",
			"<h2 id=\"header\">Header</h2>\n<p>Paragraph</p>\n",
		},
		{
			"unordered list",
			"- item 1\n- item 2",
			"<ul>\n<li>item 1</li>\n<li>item 2</li>\n</ul>\n",
		},
		{
			"ordered list",
			"1. first\n2. second",
			"<ol>\n<li>first</li>\n<li>second</li>\n</ol>\n",
		},
		{
			"inline code",
			"use `code` here",
			"<p>use <code>code</code> here</p>\n",
		},
		{
			"fenced code block",
			"```go\nfunc main() {}\n```",
			"<pre><code class=\"language-go\">func main() {}\n</code></pre>\n",
		},
		{
			"link",
			"[text](https://example.com)",
			"<p><a href=\"https://example.com\">text</a></p>\n",
		},
		{
			"table (GFM)",
			"| A | B |\n|---|---|\n| 1 | 2 |",
			"<table>\n<thead>\n<tr>\n<th>A</th>\n<th>B</th>\n</tr>\n</thead>\n<tbody>\n" +
				"<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n",
		},
		{
			"less than sign",
			"x < 5",
			"<p>x &lt; 5</p>\n",
		},
		{
			"raw HTML passthrough",
			"**bold** with <span>html</span>",
			"<p><strong>bold</strong> with <span>html</span></p>\n",
		},
		{
			"emoji",
			"✨ **sparkle** ✨",
			"<p>✨ <strong>sparkle</strong> ✨</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMarkdownToHTML(tt.input)
			if got != tt.want {
				t.Errorf("convertMarkdownToHTML(%q)\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"empty string",
			"",
			"",
		},
		{
			"safe paragraph",
			"<p>hello world</p>",
			"<p>hello world</p>",
		},
		{
			"safe bold",
			"<strong>bold</strong>",
			"<strong>bold</strong>",
		},
		{
			"safe list",
			"<ul><li>item</li></ul>",
			"<ul><li>item</li></ul>",
		},
		{
			"safe code",
			"<code>code</code>",
			"<code>code</code>",
		},
		{
			"safe link",
			"<a href=\"https://example.com\">link</a>",
			"<a href=\"https://example.com\" rel=\"nofollow\">link</a>",
		},
		{
			"script tag removed",
			"<p>text</p><script>alert('xss')</script>",
			"<p>text</p>",
		},
		{
			"iframe removed",
			"<iframe src=\"evil\"></iframe><p>text</p>",
			"<p>text</p>",
		},
		{
			"onclick removed",
			"<p onclick=\"alert('xss')\">text</p>",
			"<p>text</p>",
		},
		{
			"javascript url removed",
			"<a href=\"javascript:alert('xss')\">link</a>",
			"link",
		},
		{
			"style attribute removed",
			"<span style=\"color:red\">text</span>",
			"<span>text</span>",
		},
		{
			"safe table",
			"<table><tr><td>cell</td></tr></table>",
			"<table><tr><td>cell</td></tr></table>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeHTML(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeHTML(%q)\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrepareHTMLField(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"empty string",
			"",
			"",
		},
		{
			"plain markdown",
			"**bold** text",
			"<p><strong>bold</strong> text</p>\n",
		},
		{
			"markdown with header",
			"## Header\n\nParagraph",
			"<h2 id=\"header\">Header</h2>\n<p>Paragraph</p>\n",
		},
		{
			"markdown list",
			"- item 1\n- item 2",
			"<ul>\n<li>item 1</li>\n<li>item 2</li>\n</ul>\n",
		},
		{
			"safe HTML input",
			"<p>hello <strong>world</strong></p>",
			"<p>hello <strong>world</strong></p>",
		},
		{
			"dangerous HTML stripped",
			"<p>text</p><script>alert('xss')</script>",
			"<p>text</p>",
		},
		{
			"markdown with dangerous HTML",
			"**bold** <script>alert('xss')</script>",
			"**bold** ",
		},
		{
			"markdown with safe HTML",
			"**bold** with <span>html</span>",
			"**bold** with <span>html</span>",
		},
		{
			"HTML with onclick removed",
			"<p onclick=\"alert('xss')\">text</p>",
			"<p>text</p>",
		},
		{
			"markdown with less than",
			"x < 5 and y > 10",
			"<p>x &lt; 5 and y &gt; 10</p>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareHTMLField(tt.input)
			if got != tt.want {
				t.Errorf("prepareHTMLField(%q)\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContainsHTMLTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", false},
		{"plain text", "hello world", false},
		{"markdown bold", "**bold** text", false},
		{"markdown headers", "## Header\n\nText", false},
		{"less than sign", "x < 5 and y > 10", false},
		{"paragraph tag", "<p>text</p>", true},
		{"div tag", "<div>content</div>", true},
		{"h1 tag", "<h1>Title</h1>", true},
		{"h2 tag", "<h2>Subtitle</h2>", true},
		{"h3 tag", "<h3>Section</h3>", true},
		{"h4 tag", "<h4>Subsection</h4>", true},
		{"h5 tag", "<h5>Minor heading</h5>", true},
		{"h6 tag", "<h6>Smallest heading</h6>", true},
		{"ul tag", "<ul><li>item</li></ul>", true},
		{"ol tag", "<ol><li>item</li></ol>", true},
		{"closing tag only", "text</p>", true},
		{"mixed content", "markdown **bold** with <p>html</p>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsHTMLTags(tt.input)
			if got != tt.want {
				t.Errorf("containsHTMLTags(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
