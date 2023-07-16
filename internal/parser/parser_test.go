package parser_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/livebud/duo/internal/parser"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, name, input, expected string) {
	t.Helper()
	if name == "" {
		name = input
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		actual := parser.Print(input)
		actual = strings.ReplaceAll(actual, "  ", "")
		actual = strings.ReplaceAll(actual, "\t", "")
		actual = strings.ReplaceAll(actual, "\n", "")
		if actual == expected {
			return
		}
		var b bytes.Buffer
		b.WriteString("\n\x1b[4mInput\x1b[0m:\n")
		b.WriteString(input)
		b.WriteString("\n\n")
		b.WriteString("\x1b[4mExpected\x1b[0m:\n")
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

func equalFile(t *testing.T, name, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, name, string(input), expected)
}

func equalScope(t *testing.T, name, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	ast, err := parser.Parse(string(input))
	if err != nil {
		if err.Error() == expected {
			return
		}
	}
	actual := strings.TrimSpace(ast.Scope.String())
	expected = dedent.Dedent(expected)
	expected = strings.TrimSpace(expected)
	if actual == expected {
		return
	}
	var b bytes.Buffer
	b.WriteString("\n\x1b[4mInput\x1b[0m:\n")
	b.Write(input)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mExpected\x1b[0m:\n")
	b.WriteString(expected)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mActual\x1b[0m: \n")
	b.WriteString(actual)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mDifference\x1b[0m: \n")
	b.WriteString(diff.String(expected, actual))
	b.WriteString("\n")
	t.Fatal(b.String())
}

func Test(t *testing.T) {
	equal(t, "simple", "<h1>hi</h1>", `<h1>hi</h1>`)
	equal(t, "expr", "<h1>{greeting}</h1>", `<h1>{greeting}</h1>`)
	equal(t, "expr", "<h1>{ greeting }</h1>", `<h1>{greeting}</h1>`)
	equal(t, "self-closing", "<hr/>", `<hr />`)
	equal(t, "self-closing with space", "<hr   />", `<hr />`)
	equal(t, "attribute", `<h1 class="hello">{greeting}</h1>`, `<h1 class="hello">{greeting}</h1>`)
	equal(t, "attributes", `<h1 class="hello" id="cool">{greeting}</h1>`, `<h1 class="hello" id="cool">{greeting}</h1>`)
	equal(t, "attribute", `<hr class="hello"/>`, `<hr class="hello" />`)
	equal(t, "attribute", `<hr class="hello"   />`, `<hr class="hello" />`)
	equal(t, "attribute", `<hr data-set={set} />`, `<hr data-set="{set}" />`)
	equal(t, "attribute", `<hr class={name} />`, `<hr class="{name}" />`)
	equal(t, "attribute", `<hr class={name}/>`, `<hr class="{name}" />`)
	equal(t, "attribute", `<h1 name={greeting}>{greeting}</h1>`, `<h1 name="{greeting}">{greeting}</h1>`)
	equal(t, "attribute", `<hr class="hi-{name}-world" />`, `<hr class="hi-{name}-world" />`)
	equal(t, "attribute", `<hr class="a{b}c{d}" />`, `<hr class="a{b}c{d}" />`)
	equal(t, "attribute", `<hr {id} />`, `<hr {id} />`)
	equal(t, "attribute", `<h1 name="">{greeting}</h1>`, `<h1 name="">{greeting}</h1>`)
}

func TestEventHandler(t *testing.T) {
	equal(t, "", "<button onClick={increment}>+</button>", `<button onClick={increment}>+</button>`)
	equal(t, "", "<button onClick={() => count++}>+</button>", `<button onClick={() => { return count++; }}>+</button>`)
	equal(t, "", "<button onMouseOver={() => count++}>+</button>", `<button onMouseOver={() => { return count++; }}>+</button>`)
	equal(t, "", "<button onMouseOver={() => { count++ }}>+</button>", `<button onMouseOver={() => { count++; }}>+</button>`)
	equal(t, "", "<button onMouseOver={()=>{count++}}>+</button>", `<button onMouseOver={() => { count++; }}>+</button>`)
	equal(t, "", "<button onMouseOut={() => count++}>+</button>", `<button onMouseOut={() => { return count++; }}>+</button>`)
	equal(t, "", "<button onClick={increment} onDragStart={() => count++}>+</button>", `<button onClick={increment} onDragStart={() => { return count++; }}>+</button>`)
	equal(t, "", "<button {onClick} {onDragStart}>+</button>", `<button {onClick} {onDragStart}>+</button>`)
}

func TestIfStatement(t *testing.T) {
	equal(t, "", "{if x}{x}{end}", `{if x}{x}{end}`)
	equal(t, "", "{if x}{if y}{y}{end}{x}{end}", `{if x}{if y}{y}{end}{x}{end}`)
	equal(t, "", "{if x}\n{x}\n{end}", `{if x}{x}{end}`)
	equal(t, "", "{if x > 10}{x}{end}", `{if x > 10}{x}{end}`)
	equal(t, "", "{if (x > 10)}{x}{end}", `{if (x > 10)}{x}{end}`)
	equal(t, "", "{  if x > 10   }{  x    }{   end   }", `{if x > 10}{x}{end}`)
	equal(t, "", "{if x}{x}{else if y}{y}{end}", `{if x}{x}{else}{if y}{y}{end}{end}`)
	equal(t, "", "{if x}{x}{else if (y)}{y}{end}", `{if x}{x}{else}{if (y)}{y}{end}{end}`)
	equal(t, "", "{if x}\n{x}\n{else if y}\n{y}\n{end}", `{if x}{x}{else}{if y}{y}{end}{end}`)
	equal(t, "", "{   if x   }{x}{    else if y  }{y}{   end  }", `{if x}{x}{else}{if y}{y}{end}{end}`)
	equal(t, "", "{if x == 10}{x}{else if y > 10}{y}{end}", `{if x == 10}{x}{else}{if y > 10}{y}{end}{end}`)
	equal(t, "", "{if x == 10}{x}{else if (y > 10)}{y}{end}", `{if x == 10}{x}{else}{if (y > 10)}{y}{end}{end}`)
	equal(t, "", "{if x == 10}{x}{else if y > 10}{y}{else}none{end}", `{if x == 10}{x}{else}{if y > 10}{y}{else}none{end}{end}`)
	equal(t, "", "{  if     x   ==   10  }{  x  }{   else    if    y > 10   }{  y   }{   else   }none{   end   }", `{if x == 10}{x}{else}{if y > 10}{y}{else}none{end}{end}`)
	equal(t, "", "{if x}{x}{else   }{y}{end}", `{if x}{x}{else}{y}{end}`)
	equal(t, "", "{if x}{x}{else}{y}{end}", `{if x}{x}{else}{y}{end}`)
}

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.html", `<script>export let greeting = "hello"; setInterval(() => { greeting += "o"; }, 500); </script><h1>{greeting}</h1>`)
	equalFile(t, "02-attribute.html", `<div><hr {name} /><hr name="{name}" /><hr name="{name}" /><hr name="{target}-{name}" /><hr name="" /></div>`)
	equalFile(t, "03-counter.html", `<script>export let count = 0; function increment () { count += 1; }; </script><button onClick={increment}>Clicked {count || 0} {count === 1 ? 'time' : 'times'}</button>`)
}

func TestScope(t *testing.T) {
	equalScope(t, "01-greeting.html", `
		"greeting" declared=true exported=true mutable=true
		"setInterval" declared=false exported=false mutable=false
	`)
	equalScope(t, "02-attribute.html", `
		"name" declared=false exported=false mutable=false
		"target" declared=false exported=false mutable=false
	`)
	equalScope(t, "03-counter.html", `
		"count" declared=true exported=true mutable=true
		"increment" declared=true exported=false mutable=false
	`)
}
