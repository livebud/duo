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
	equal(t, "attribute", `<button onclick={addToList} disabled={newItem === ""}>Add</button>`, `<button onclick="{addToList}" disabled="{newItem === ""}">Add</button>`)
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
	equal(t, "", "{#if x}{x}{/if}", `{#if x}{x}{/if}`)
	equal(t, "", "{#if x}\n{x}\n{/if}", `{#if x}{x}{/if}`)
	equal(t, "", "{#if x > 10}{x}{/if}", `{#if x > 10}{x}{/if}`)
	equal(t, "", "{#if x > 10}{#if (y == 10)}{x}{/if}{/if}", `{#if x > 10}{#if (y == 10)}{x}{/if}{/if}`)
	equal(t, "", "{#if (x > 10)}{x}{/if}", `{#if (x > 10)}{x}{/if}`)
	equal(t, "", "{  #if   x   >   10   }{  x   }{   /if   }", `{#if x > 10}{x}{/if}`)
	equal(t, "", "{#if x}{x}{:else if y}{y}{/if}", `{#if x}{x}{:else}{#if y}{y}{/if}{/if}`)
	equal(t, "", "{#if x}{x}{:else if (y)}{y}{/if}", `{#if x}{x}{:else}{#if (y)}{y}{/if}{/if}`)
	equal(t, "", "{#if x}\n{x}\n{:else if y}\n{y}\n{/if}", `{#if x}{x}{:else}{#if y}{y}{/if}{/if}`)
	equal(t, "", "{   #if x   }{x}{    :else    if y  }{y}{   /if  }", `{#if x}{x}{:else}{#if y}{y}{/if}{/if}`)
	equal(t, "", "{#if x == 10}{x}{:else if y > 10}{y}{/if}", `{#if x == 10}{x}{:else}{#if y > 10}{y}{/if}{/if}`)
	equal(t, "", "{#if x == 10}{x}{:else if (y > 10)}{y}{/if}", `{#if x == 10}{x}{:else}{#if (y > 10)}{y}{/if}{/if}`)
	equal(t, "", "{#if x == 10}{x}{:else if y > 10}{y}{:else}none{/if}", `{#if x == 10}{x}{:else}{#if y > 10}{y}{:else}none{/if}{/if}`)
	equal(t, "", "{  #if     x   ==   10  }{  x  }{   :else    if    y > 10   }{  y   }{   :else   }none{   /if   }", `{#if x == 10}{x}{:else}{#if y > 10}{y}{:else}none{/if}{/if}`)
	equal(t, "", "{#if x}{x}{:else}{y}{/if}", `{#if x}{x}{:else}{y}{/if}`)
	equal(t, "", "{#if x}{x}{  :else  }{y}{/if}", `{#if x}{x}{:else}{y}{/if}`)
	equal(t, "", "<h1>{#if greeting}hi{:else if planet}mars{/if}</h1>", `<h1>{#if greeting}hi{:else}{#if planet}mars{/if}{/if}</h1>`)
}

func TestEachLoop(t *testing.T) {
	equal(t, "", "{#each items as item}{item}{/each}", `{#each items as item}{item}{/each}`)
	equal(t, "", "{#each items as item}\n{item}\n{/each}", `{#each items as item}{item}{/each}`)
	equal(t, "", "{#each   items    as   item}  \n  {  item  }  \n  { / each  }", `{#each items as item}{item}{/each}`)
	equal(t, "", "{#each items as item, i}{i}:{item}{/each}", `{#each items as item, i}{i}:{item}{/each}`)
	equal(t, "", "{#each  items   as   i, item}\n{i}:{item}\n{/each}", `{#each items as i, item}{i}:{item}{/each}`)
	equal(t, "", "{#each   items  as      item  ,  i   }  \n  {  i  }:{  item  }\n{ / each  }", `{#each items as item, i}{i}:{item}{/each}`)
	equal(t, "", "{#each items as 3}{3}{/each}", `parser: {#each items as 3}{3}{/each}: expected an identifier, got *js.LiteralExpr`)
	equal(t, "", "{#each items}{outer}{/each}", `{#each items}{outer}{/each}`)
	equal(t, "", "{#each   items  }{outer}{/each}", `{#each items}{outer}{/each}`)
	// TODO: handle parsing destructured objects
	// equal(t, "", "{#each cats as { id, name }, i}{id}:{name}{/each}", ``)
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

func TestBind(t *testing.T) {
	equal(t, "", `<input type="text" bind:value={name} />`, `<input type="text" bind:value={name} />`)
	equal(t, "", `<input bind:value={todo.newItem} type="text" placeholder="new todo item.." />`, `<input bind:value={todo.newItem} type="text" placeholder="new todo item.." />`)
}

func TestClass(t *testing.T) {
	equal(t, "", `<span class:checked={item.status}>{item.text}</span>`, `<span class:checked={item.status}>{item.text}</span>`)
}

func TestStyle(t *testing.T) {
	equal(t, "", `<span>{item.text}</span><style>span { background-color: blue; }</style>`, `<span>{item.text}</span><style>span { background-color: blue }</style>`)
}
