package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/livebud/duo/internal/lexer"
	"github.com/livebud/duo/internal/parser"
	"github.com/livebud/duo/internal/token"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, path, input, expected string) {
	t.Helper()
	if path == "" {
		path = input
	}
	t.Run(path, func(t *testing.T) {
		t.Helper()
		actual := parser.Print(path, input)
		actual = strings.ReplaceAll(actual, "  ", "")
		actual = strings.ReplaceAll(actual, "\t", "")
		actual = strings.ReplaceAll(actual, "\n", "")
		diff.TestString(t, actual, expected)
	})
}

func equalFile(t *testing.T, path, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", path))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, path, string(input), expected)
}

func equalScope(t *testing.T, path, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", path))
	if err != nil {
		t.Fatal(err)
	}
	ast, err := parser.Parse(path, string(input))
	if err != nil {
		if err.Error() == expected {
			return
		}
		t.Fatal(err)
	}
	actual := strings.TrimSpace(ast.Scope.String())
	expected = dedent.Dedent(expected)
	expected = strings.TrimSpace(expected)
	diff.TestString(t, actual, expected)
}

func TestAPI(t *testing.T) {
	is := is.New(t)
	p := parser.New("", lexer.New("<h1>hi</h1>"))
	is.Equal(p.Accept(token.Text), false)
	is.Equal(p.Type(), token.Type(""))
	is.Equal(p.Text(), "")
	is.Equal(p.Accept(token.LessThan), true)
	is.Equal(p.Type(), token.LessThan)
	is.Equal(p.Text(), "<")
	is.Equal(p.Accept(token.Identifier), true)
	is.Equal(p.Type(), token.Identifier)
	is.Equal(p.Text(), "h1")
	// Keep old if accept if false
	is.Equal(p.Accept(token.If), false)
	is.Equal(p.Type(), token.Identifier)
	is.Equal(p.Text(), "h1")
	is.Equal(p.Is(token.GreaterThan), true)
	is.Equal(p.Accept(token.GreaterThan), true)
	is.Equal(p.Type(), token.GreaterThan)
	is.Equal(p.Text(), ">")
	is.Equal(p.Accept(token.Text), true)
	is.Equal(p.Type(), token.Text)
	is.Equal(p.Text(), "hi")
	is.Equal(p.Accept(token.LessThanSlash), true)
	is.Equal(p.Type(), token.LessThanSlash)
	is.Equal(p.Text(), "</")
	is.Equal(p.Accept(token.Identifier), true)
	is.Equal(p.Type(), token.Identifier)
	is.Equal(p.Text(), "h1")
	is.Equal(p.Accept(token.GreaterThan), true)
	is.Equal(p.Type(), token.GreaterThan)
	is.Equal(p.Text(), ">")
	is.Equal(p.Accept(token.Identifier), false)
	is.Equal(p.Type(), token.GreaterThan)
	is.Equal(p.Text(), ">")
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
	equal(t, "attribute", `<hr class={"name"}/>`, `<hr class="{"name"}" />`)
	equal(t, "attribute", `<h1 name={greeting}>{greeting}</h1>`, `<h1 name="{greeting}">{greeting}</h1>`)
	equal(t, "attribute", `<hr class="hi-{name}-world" />`, `<hr class="hi-{name}-world" />`)
	equal(t, "attribute", `<hr class="a{b}c{d}" />`, `<hr class="a{b}c{d}" />`)
	equal(t, "attribute", `<hr {id} />`, `<hr {id} />`)
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

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.html", `<script>export let greeting = "hello"; setInterval(() => { greeting += "o"; }, 500); </script><h1>{greeting}</h1>`)
	equalFile(t, "02-attribute.html", `<div><hr {name} /><hr name="{name}" /><hr name="{name}" /><hr name="{target}-{name}" /><hr name="" /></div>`)
	equalFile(t, "03-counter.html", `<script>export let count = 0; function increment () { count += 1; }; </script><button onClick={increment}>Clicked {count || 0} {count === 1 ? 'time' : 'times'}</button>`)
}

func TestScope(t *testing.T) {
	equalScope(t, "01-greeting.html", `
		"greeting" declared exported mutable
		"setInterval"
	`)
	equalScope(t, "02-attribute.html", `
		"name"
		"target"
	`)
	equalScope(t, "03-counter.html", `
		"count" declared exported mutable
		"increment" declared
	`)
	equalScope(t, "04-component.html", `
		"Sub" import="./04-sub.html" default
	`)
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
	equal(t, "", "<h1>{if greeting}hi{else if planet}mars{end}</h1>", `<h1>{if greeting}hi{else}{if planet}mars{end}{end}</h1>`)
	equal(t, "", "<h1>{if greeting}hi{else if planet}mars{else}world{end}</h1>", `<h1>{if greeting}hi{else}{if planet}mars{else}world{end}{end}</h1>`)
	equal(t, "", "<h1>{if greeting}hi{else if planet}mars{else if name}world{end}</h1>", `<h1>{if greeting}hi{else}{if planet}mars{else}{if name}world{end}{end}{end}</h1>`)
	equal(t, "", "<h1>{if greeting}hi{else if planet}mars{else if name}world{else}universe{end}</h1>", `<h1>{if greeting}hi{else}{if planet}mars{else}{if name}world{else}universe{end}{end}{end}</h1>`)
}

func TestForLoop(t *testing.T) {
	equal(t, "", "{for item in items}{item}{end}", `{for item in items}{item}{end}`)
	equal(t, "", "{for item in items}\n{item}\n{end}", `{for item in items}{item}{end}`)
	equal(t, "", "{for   item    in   items}  \n  {  item  }  \n  {  end  }", `{for item in items}{item}{end}`)
	equal(t, "", "{for i, item in items}{i}:{item}{end}", `{for i, item in items}{i}:{item}{end}`)
	equal(t, "", "{for i, item in items}\n{i}:{item}\n{end}", `{for i, item in items}{i}:{item}{end}`)
	equal(t, "", "{for   i  ,   item   in   items  }  \n  {  i  }:{  item  }\n{  end  }", `{for i, item in items}{i}:{item}{end}`)
	equal(t, "", "{for 3 in items}{3}{end}", `unexpected token '3'`)
	equal(t, "", "{for items}{item}{end}", `{for items}{item}{end}`)
	equal(t, "", "{for   items  }{item}{end}", `{for items}{item}{end}`)
}

func TestComponent(t *testing.T) {
	equal(t, "", "<Component/>", `<Component />`)
	equal(t, "", "<Component></Component>", `<Component></Component>`)
	equal(t, "", "<FirstName/>", `<FirstName />`)
	equal(t, "", "<FirstName></FirstName>", `<FirstName></FirstName>`)
	equal(t, "", "<H1/>", `<H1 />`)
	equal(t, "", "<H1>hi</H1>", `<H1>hi</H1>`)
	equal(t, "", "<Component a={b} />", `<Component a="{b}" />`)
	equal(t, "", "<FirstName {props} />", `<FirstName {props} />`)
	equal(t, "", `<script>import Component from "./Component.duo";</script><Component/>`, `<script>import Component from "./Component.duo"; </script><Component />`)
	equal(t, "", `<script>import Component from "./component.duo";</script><Component/>`, `<script>import Component from "./component.duo"; </script><Component />`)
	equal(t, "", `<script>import Component from "./component.duo";</script><Component a={b}/>`, `<script>import Component from "./component.duo"; </script><Component a="{b}" />`)
	equal(t, "", `<script>import A from "./a.duo"; import B from "./b.duo";</script><A/><B/>`, `<script>import A from "./a.duo"; import B from "./b.duo"; </script><A /><B />`)
	equal(t, "", `<script>import A from "./a.duo"; import B from "./b.duo"; import C from './c.duo';</script><A/><B/>`, `<script>import A from "./a.duo"; import B from "./b.duo"; import C from "./c.duo"; </script><A /><B />`)
}

func TestSlot(t *testing.T) {
	equal(t, "", "<slot/>", `<slot />`)
	equal(t, "", "<slot  />", `<slot />`)
	equal(t, "", "<slot  >fallback</slot>", `<slot>fallback</slot>`)
	equal(t, "", "<slot name=\"value\" ><span>1</span><span>2</span></slot>", `<slot name="value"><span>1</span><span>2</span></slot>`)
	equal(t, "", "<span slot=\"name\">fallback</span>", `<span slot="name">fallback</span>`)
}

func TestScript(t *testing.T) {
	equal(t, "", "<script>let posts = [];</script>", `<script>let posts = []; </script>`)
	equal(t, "", "<script>let posts: Post[] = [];</script>", `<script>let posts = []; </script>`)
}

func TestComment(t *testing.T) {
	equal(t, "", "<!-- this is a comment -->\n<h2>hello world</h2>", "<!-- this is a comment --><h2>hello world</h2>")
}
