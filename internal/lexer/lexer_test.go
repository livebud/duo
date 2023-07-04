package lexer_test

import (
	"bytes"
	"testing"

	"github.com/livebud/duo/internal/lexer"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, name, input, expected string) {
	t.Helper()
	if name == "" {
		name = input
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		actual := lexer.Print(input)
		if actual == expected {
			return
		}
		var b bytes.Buffer
		b.WriteString("\n\x1b[4mInput\x1b[0m:\n")
		b.WriteString(input)
		b.WriteString("\n\x1b[4mExpected\x1b[0m:\n")
		b.WriteString(expected)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mActual\x1b[0m: \n")
		b.WriteString(actual)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mDifference\x1b[0m: \n")
		b.WriteString(diff.String(expected, actual))
		b.WriteString("\n")
		t.Fatal(b.String())
	})
}

func TestHTML(t *testing.T) {
	equal(t, "simple", "<h1>hi</h1>", `< identifier:"h1" > text:"hi" </ identifier:"h1" >`)
	equal(t, "text", "hello", `text:"hello"`)
	equal(t, "selfclosing", "<br/>", `< identifier:"br" />`)
	equal(t, "", "<p>Paragraph</p>", `< identifier:"p" > text:"Paragraph" </ identifier:"p" >`)
	equal(t, "", "<div><span>Text</span></div>", `< identifier:"div" > < identifier:"span" > text:"Text" </ identifier:"span" > </ identifier:"div" >`)
	equal(t, "", "<a href=\"https://example.com\">Link</a>", `< identifier:"a" space:" " identifier:"href" = quote:"\"" text:"https://example.com" quote:"\"" > text:"Link" </ identifier:"a" >`)
	equal(t, "", "<img src=\"image.jpg\" alt=\"Image\">", `< identifier:"img" space:" " identifier:"src" = quote:"\"" text:"image.jpg" quote:"\"" space:" " identifier:"alt" = quote:"\"" text:"Image" quote:"\"" >`)
	equal(t, "", "<input type=\"text\" value=\"John Doe\">", `< identifier:"input" space:" " identifier:"type" = quote:"\"" text quote:"\"" space:" " identifier:"value" = quote:"\"" text:"John Doe" quote:"\"" >`)
	equal(t, "", "<button disabled>Click me</button>", `< identifier:"button" space:" " identifier:"disabled" > text:"Click me" </ identifier:"button" >`)
	equal(t, "", "<ul><li>Item 1</li><li>Item 2</li></ul>", `< identifier:"ul" > < identifier:"li" > text:"Item 1" </ identifier:"li" > < identifier:"li" > text:"Item 2" </ identifier:"li" > </ identifier:"ul" >`)
	equal(t, "form", "<form action=\"/submit\" method=\"post\"><input type=\"text\" name=\"name\"/><input type=\"submit\" value=\"Submit\"></form>", `< identifier:"form" space:" " identifier:"action" = quote:"\"" text:"/submit" quote:"\"" space:" " identifier:"method" = quote:"\"" text:"post" quote:"\"" > < identifier:"input" space:" " identifier:"type" = quote:"\"" text quote:"\"" space:" " identifier:"name" = quote:"\"" text:"name" quote:"\"" /> < identifier:"input" space:" " identifier:"type" = quote:"\"" text:"submit" quote:"\"" space:" " identifier:"value" = quote:"\"" text:"Submit" quote:"\"" > </ identifier:"form" >`)
	equal(t, "", "<br>", `< identifier:"br" >`)
	equal(t, "", "<br/>", `< identifier:"br" />`)
	equal(t, "", "<br />", `< identifier:"br" space:" " />`)
	equal(t, "", "<div class=\"container\"><h2>Title</h2><p>Content</p></div>", `< identifier:"div" space:" " identifier:"class" = quote:"\"" text:"container" quote:"\"" > < identifier:"h2" > text:"Title" </ identifier:"h2" > < identifier:"p" > text:"Content" </ identifier:"p" > </ identifier:"div" >`)
	equal(t, "", "<div id=4", `< identifier:"div" space:" " identifier:"id" = text:"4" error:"unexpected end of input"`)
	equal(t, "", "<script>alert('Hello, world!');</script>", `< script > text:"alert('Hello, world!');" </ script >`)
	equal(t, "", "<script>alert('Hello, world!');</script><h1>hello</h1>", `< script > text:"alert('Hello, world!');" </ script > < identifier:"h1" > text:"hello" </ identifier:"h1" >`)
	equal(t, "", "<script>alert('Hello,<h1> world!');</script>", `< script > text:"alert('Hello,<h1> world!');" </ script >`)
	equal(t, "", "<script>alert('Hello,<h1> world!');</script>", `< script > text:"alert('Hello,<h1> world!');" </ script >`)
	equal(t, "", "<style>body { font-family: Arial; }</style><h1>hello</h1>", `< style > text:"body { font-family: Arial; }" </ style > < identifier:"h1" > text:"hello" </ identifier:"h1" >`)
	equal(t, "", "<style>body <h1> { font-family: Arial; }</style>", `< style > text:"body <h1> { font-family: Arial; }" </ style >`)
	equal(t, "", "<table><tr><td>Cell 1</td><td>Cell 2</td></tr></table>", `< identifier:"table" > < identifier:"tr" > < identifier:"td" > text:"Cell 1" </ identifier:"td" > < identifier:"td" > text:"Cell 2" </ identifier:"td" > </ identifier:"tr" > </ identifier:"table" >`)
	equal(t, "", "<pre>Formatted text</pre>", `< identifier:"pre" > text:"Formatted text" </ identifier:"pre" >`)
	equal(t, "", "<strong>Bold text</strong>", `< identifier:"strong" > text:"Bold text" </ identifier:"strong" >`)
	equal(t, "", "<em>Italic text</em>", `< identifier:"em" > text:"Italic text" </ identifier:"em" >`)
	equal(t, "", "<blockquote>Quote</blockquote>", `< identifier:"blockquote" > text:"Quote" </ identifier:"blockquote" >`)
	equal(t, "", "<code>code()</code>", `< identifier:"code" > text:"code()" </ identifier:"code" >`)
	equal(t, "", "<ol><li>Item 1</li><li>Item 2</li></ol>", `< identifier:"ol" > < identifier:"li" > text:"Item 1" </ identifier:"li" > < identifier:"li" > text:"Item 2" </ identifier:"li" > </ identifier:"ol" >`)
	equal(t, "headers", "<h1>Title 1</h1><h2>Title 2</h2><h3>Title 3</h3><h4>Title 4</h4><h5>Title 5</h5><h6>Title 6</h6><h7>Title 6</h7>", `< identifier:"h1" > text:"Title 1" </ identifier:"h1" > < identifier:"h2" > text:"Title 2" </ identifier:"h2" > < identifier:"h3" > text:"Title 3" </ identifier:"h3" > < identifier:"h4" > text:"Title 4" </ identifier:"h4" > < identifier:"h5" > text:"Title 5" </ identifier:"h5" > < identifier:"h6" > text:"Title 6" </ identifier:"h6" > < identifier:"h7" > text:"Title 6" </ identifier:"h7" >`)
	equal(t, "", "<hr>", `< identifier:"hr" >`)
	equal(t, "", "<hr/>", `< identifier:"hr" />`)
	equal(t, "", "<audio src=\"music.mp3\" controls></audio>", `< identifier:"audio" space:" " identifier:"src" = quote:"\"" text:"music.mp3" quote:"\"" space:" " identifier:"controls" > </ identifier:"audio" >`)
	equal(t, "", "<video src=\"video.mp4\" controls></video>", `< identifier:"video" space:" " identifier:"src" = quote:"\"" text:"video.mp4" quote:"\"" space:" " identifier:"controls" > </ identifier:"video" >`)
	equal(t, "", "<canvas width=\"300\" height=\"200\"></canvas>", `< identifier:"canvas" space:" " identifier:"width" = quote:"\"" text:"300" quote:"\"" space:" " identifier:"height" = quote:"\"" text:"200" quote:"\"" > </ identifier:"canvas" >`)
	equal(t, "", "<canvas width=300 height=200></canvas>", `< identifier:"canvas" space:" " identifier:"width" = text:"300" space:" " identifier:"height" = text:"200" > </ identifier:"canvas" >`)
	equal(t, "", "<svg width=\"100\" height=\"100\"><circle cx=\"50\" cy=\"50\" r=\"40\"></circle></svg>", `< identifier:"svg" space:" " identifier:"width" = quote:"\"" text:"100" quote:"\"" space:" " identifier:"height" = quote:"\"" text:"100" quote:"\"" > < identifier:"circle" space:" " identifier:"cx" = quote:"\"" text:"50" quote:"\"" space:" " identifier:"cy" = quote:"\"" text:"50" quote:"\"" space:" " identifier:"r" = quote:"\"" text:"40" quote:"\"" > </ identifier:"circle" > </ identifier:"svg" >`)
	equal(t, "", "<div style=\"color: <h1>red;\">Red text</div>", `< identifier:"div" space:" " identifier:"style" = quote:"\"" text:"color: <h1>red;" quote:"\"" > text:"Red text" </ identifier:"div" >`)
	equal(t, "", "<a href=\"#section\">Jump to section</a>", `< identifier:"a" space:" " identifier:"href" = quote:"\"" text:"#section" quote:"\"" > text:"Jump to section" </ identifier:"a" >`)
	equal(t, "", "<iframe src=\"https://example.com\" width=\"500\" height=\"300\"></iframe>", `< identifier:"iframe" space:" " identifier:"src" = quote:"\"" text:"https://example.com" quote:"\"" space:" " identifier:"width" = quote:"\"" text:"500" quote:"\"" space:" " identifier:"height" = quote:"\"" text:"300" quote:"\"" > </ identifier:"iframe" >`)
	equal(t, "", "<meta charset=\"UTF-8\" />", `< identifier:"meta" space:" " identifier:"charset" = quote:"\"" text:"UTF-8" quote:"\"" space:" " />`)
}

func TestComment(t *testing.T) {
	equal(t, "", "<!-- Comment -->", `comment:"<!-- Comment -->"`)
	equal(t, "", "<h1><!-- Comment --></h1>", `< identifier:"h1" > comment:"<!-- Comment -->" </ identifier:"h1" >`)
	equal(t, "", "<h1>hi<!-- Comment -->world</h1>", `< identifier:"h1" > text:"hi" comment:"<!-- Comment -->" text:"world" </ identifier:"h1" >`)
	equal(t, "", "<h1/><!-- Comment -->", `< identifier:"h1" /> comment:"<!-- Comment -->"`)
	equal(t, "", "<!-- Comment --><h1/>", `comment:"<!-- Comment -->" < identifier:"h1" />`)
	equal(t, "", "<h1 <!-- Comment -->>", `< identifier:"h1" space:" " error:"unexpected token '<'" text:"!-- Comment -->>"`)
}

func TestExpression(t *testing.T) {
	equal(t, "expression", "<h1>{greeting}</h1>", `< identifier:"h1" > { expr:"greeting" } </ identifier:"h1" >`)
	equal(t, "expression", "<h1>{greeting && session}</h1>", `< identifier:"h1" > { expr:"greeting && session" } </ identifier:"h1" >`)
	equal(t, "expression", "<h1>{greeting && </h1>}</h1>", `< identifier:"h1" > { expr:"greeting && </h1>" } </ identifier:"h1" >`)
	equal(t, "expression", "<h1>hello {planet}!</h1>", `< identifier:"h1" > text:"hello " { expr:"planet" } text:"!" </ identifier:"h1" >`)
	equal(t, "expression", "<h1>{greeting && \"</h1>\"}</h1>", `< identifier:"h1" > { expr:"greeting && \"</h1>\"" } </ identifier:"h1" >`)
}

func TestDoctype(t *testing.T) {
	equal(t, "", "<!doctype html>", `<!doctype space:" " identifier:"html" >`)
	equal(t, "", "<!doctype html/>", `<!doctype space:" " identifier:"html" />`)
	equal(t, "", "<!doctype html />", `<!doctype space:" " identifier:"html" space:" " />`)
	equal(t, "", "<!DOCTYPE html>", `<!doctype:"<!DOCTYPE" space:" " identifier:"html" >`)
	equal(t, "", "<!DOCTYPE html>", `<!doctype:"<!DOCTYPE" space:" " identifier:"html" >`)
	equal(t, "", "<!DOCTYPE html />", `<!doctype:"<!DOCTYPE" space:" " identifier:"html" space:" " />`)
	equal(t, "", "<!DOCTYPE html/>", `<!doctype:"<!DOCTYPE" space:" " identifier:"html" />`)
}
