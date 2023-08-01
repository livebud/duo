package evaluator_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/livebud/duo/internal/evaluator"
	"github.com/livebud/duo/internal/resolver"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, path, input string, props interface{}, expected string) {
	t.Helper()
	if path == "" {
		path = input
	}
	t.Run(path, func(t *testing.T) {
		t.Helper()
		resolver := resolver.Embedded{
			path: []byte(input),
		}
		evaluator := evaluator.New(resolver)
		str := new(strings.Builder)
		actual := ""
		if err := evaluator.Evaluate(str, path, props); err != nil {
			actual = err.Error()
		} else {
			actual = str.String()
		}
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

func equalMap(t *testing.T, files map[string]string, props interface{}, expected string) {
	t.Helper()
	input := files["main.duo"]
	if input == "" {
		t.Fatal("missing main.duo")
	}
	t.Run(input, func(t *testing.T) {
		t.Helper()
		str := new(strings.Builder)
		actual := ""
		resolver := resolver.Embedded{}
		for path, code := range files {
			resolver[path] = []byte(code)
		}
		evaluator := evaluator.New(resolver)
		if err := evaluator.Evaluate(str, "main.duo", props); err != nil {
			actual = err.Error()
		} else {
			actual = str.String()
		}
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

func equalFile(t *testing.T, name string, props interface{}, expected string) {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, name, strings.TrimSpace(string(input)), props, expected)
}

type Map = map[string]interface{}

func TestSimple(t *testing.T) {
	equal(t, "simple", `<h1>hi</h1>`, nil, `<h1>hi</h1>`)
	equal(t, "greeting", `<h1>{greeting}</h1>`, nil, `<h1></h1>`)
	equal(t, "greeting", `<h1>{greeting}</h1>`, Map{"greeting": "hi"}, `<h1>hi</h1>`)
	equal(t, "greeting", `<h1>{greeting}</h1>`, Map{}, `<h1></h1>`)
	equal(t, "planet", `<h1>hello {planet}!</h1>`, Map{"planet": "mars"}, `<h1>hello mars!</h1>`)
	equal(t, "planet", `<h1>hello {planet}!</h1>`, Map{}, `<h1>hello !</h1>`)
	equal(t, "greeting_planet", `<h1>{greeting} {planet}!</h1>`, Map{"greeting": "hola", "planet": "Earth"}, `<h1>hola Earth!</h1>`)
	equal(t, "attributes", `<hr target={target} />`, Map{"target": "_blank"}, `<hr target="_blank"/>`)
	equal(t, "attributes", `<hr target={target} />`, Map{}, `<hr/>`)
	equal(t, "attributes", `<hr name="{target}-{name}" />`, Map{"target": "_blank", "name": "anki"}, `<hr name="_blank-anki"/>`)
	equal(t, "attributes", `<hr name="{target}-{name}" />`, Map{"target": "_blank"}, `<hr name="_blank-"/>`)
	equal(t, "attributes", `<hr name="{target}-{name}" />`, Map{"name": "anki"}, `<hr name="-anki"/>`)
	equal(t, "attributes", `<hr target="{target}-{name}" />`, Map{}, `<hr target="-"/>`)
	equal(t, "attributes", `<hr {name} />`, Map{"name": "hello"}, `<hr name="hello"/>`)
	equal(t, "attributes", `<hr {name} />`, Map{}, `<hr/>`)
	// TODO: Should this be `<h1 name></h1>`?
	equal(t, "attributes", `<h1 name=""></h1>`, Map{}, `<h1></h1>`)
}

func TestEventHandler(t *testing.T) {
	equal(t, "", "<button onClick={increment}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button onClick={() => count++}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button onMouseOver={() => count++}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button onMouseOver={() => { count++ }}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button onMouseOver={()=>{count++}}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button onMouseOut={() => count++}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button onClick={increment} onDragStart={() => count++}>+</button>", Map{}, `<button>+</button>`)
	equal(t, "", "<button {onClick} {onDragStart}>+</button>", Map{}, `<button>+</button>`)
}

func TestExpr(t *testing.T) {
	equal(t, "", `<h1>{1+1}</h1>`, Map{}, `<h1>2</h1>`)
	equal(t, "", `<h1>{true?'a':'b'}</h1>`, Map{}, `<h1>a</h1>`)
	equal(t, "", `<h1>{false?'a':'b'}</h1>`, Map{}, `<h1>b</h1>`)
	equal(t, "", `<h1>{true===true?'a':'b'}</h1>`, Map{}, `<h1>a</h1>`)
	equal(t, "", `<h1>{true===false?'a':'b'}</h1>`, Map{}, `<h1>b</h1>`)
	equal(t, "", `<button>{count} {count === 1 ? 'time' : 'times'}</button>`, Map{}, `<button> times</button>`)
	equal(t, "", `<button>{count} {count === 1 ? 'time' : 'times'}</button>`, Map{"count": 0}, `<button>0 times</button>`)
	equal(t, "", `<button>{count} {count === 1 ? 'time' : 'times'}</button>`, Map{"count": 1}, `<button>1 time</button>`)
	equal(t, "", `<button>{count} {count === 1 ? 'time' : 'times'}</button>`, Map{"count": 2}, `<button>2 times</button>`)
	equal(t, "", `<button>{count} {count === 1 ? 'time' : 'times'}</button>`, Map{"count": 99}, `<button>99 times</button>`)
}

func TestFile(t *testing.T) {
	equalFile(t, "01-greeting.html", Map{}, "\n\n<h1></h1>")
	equalFile(t, "01-greeting.html", Map{"greeting": "hi"}, "\n\n<h1>hi</h1>")
	equalFile(t, "02-attribute.html", Map{}, "<div>\n  <hr/>\n  <hr/>\n  <hr/>\n  <hr name=\"-\"/>\n  <hr/>\n</div>")
	equalFile(t, "02-attribute.html", Map{"name": "anki"}, "<div>\n  <hr name=\"anki\"/>\n  <hr name=\"anki\"/>\n  <hr name=\"anki\"/>\n  <hr name=\"-anki\"/>\n  <hr/>\n</div>")
	equalFile(t, "02-attribute.html", Map{"target": "window", "name": "anki"}, "<div>\n  <hr name=\"anki\"/>\n  <hr name=\"anki\"/>\n  <hr name=\"anki\"/>\n  <hr name=\"window-anki\"/>\n  <hr/>\n</div>")
	equalFile(t, "03-counter.html", Map{}, "\n\n<button>\n  Clicked 0 times\n</button>")
	equalFile(t, "03-counter.html", Map{"count": 1}, "\n\n<button>\n  Clicked 1 time\n</button>")
	equalFile(t, "03-counter.html", Map{"count": 10}, "\n\n<button>\n  Clicked 10 times\n</button>")
}

func TestIf(t *testing.T) {
	equal(t, "", `<h1>{if greeting}hi{end}</h1>`, Map{}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{end}</h1>`, Map{"greeting": true}, `<h1>hi</h1>`)
	equal(t, "", `<h1>{if greeting}hi{end}</h1>`, Map{"greeting": false}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else}bye{end}</h1>`, Map{}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else}bye{end}</h1>`, Map{"greeting": true}, `<h1>hi</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else}bye{end}</h1>`, Map{"greeting": false}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{end}</h1>`, Map{}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{end}</h1>`, Map{"greeting": true}, `<h1>hi</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{end}</h1>`, Map{"greeting": false}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{end}</h1>`, Map{"planet": true}, `<h1>mars</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{end}</h1>`, Map{"planet": false}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else}bye{end}</h1>`, Map{}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else}bye{end}</h1>`, Map{"greeting": true}, `<h1>hi</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else}bye{end}</h1>`, Map{"greeting": false}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else}bye{end}</h1>`, Map{"planet": true}, `<h1>mars</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else}bye{end}</h1>`, Map{"planet": false}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{"greeting": true}, `<h1>hi</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{"greeting": false}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{"planet": true}, `<h1>mars</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{"planet": false}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{"name": true}, `<h1>anki</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{end}</h1>`, Map{"name": false}, `<h1></h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{"greeting": true}, `<h1>hi</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{"greeting": false}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{"planet": true}, `<h1>mars</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{"planet": false}, `<h1>bye</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{"name": true}, `<h1>anki</h1>`)
	equal(t, "", `<h1>{if greeting}hi{else if planet}mars{else if name}anki{else}bye{end}</h1>`, Map{"name": false}, `<h1>bye</h1>`)
}

func TestFor(t *testing.T) {
	equal(t, "", `<ul>{for item in items}<li>{item}</li>{end}</ul>`, Map{"items": []string{"a", "b", "c"}}, `<ul><li>a</li><li>b</li><li>c</li></ul>`)
	equal(t, "", `<ul>{for item in items}<li>{item}</li>{end}</ul>`, Map{"items": []string{}}, `<ul></ul>`)
	equal(t, "", `<ul>{for item in items}<li>{item}</li>{end}</ul>`, Map{}, `<ul></ul>`)
	equal(t, "", `<ul>{for i, item in items}<li>{i}. {item}</li>{end}</ul>`, Map{"items": []string{"a", "b", "c"}}, `<ul><li>0. a</li><li>1. b</li><li>2. c</li></ul>`)
	equal(t, "", `<ul>{for i, item in items}<li>{i}. {item}</li>{end}</ul>`, Map{"items": []string{}}, `<ul></ul>`)
	equal(t, "", `<ul>{for i, item in items}<li>{i}. {item}</li>{end}</ul>`, Map{}, `<ul></ul>`)
}

func TestComponent(t *testing.T) {
	equalMap(t, map[string]string{
		"Component.duo": `<h1>Component</h1>`,
		"main.duo":      `<script>import Component from "./Component.duo";</script><Component/>`,
	}, Map{}, `<h1>Component</h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>Component</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component/>`,
	}, Map{}, `<h1>Component</h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>{title}</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component title="hello"/>`,
	}, Map{}, `<h1>hello</h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>{title}</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component/>`,
	}, Map{"title": "hello"}, `<h1></h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>{title}</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component />`,
	}, Map{}, `<h1></h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>{title}</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component title={h1}></Component><Component title={h2}/>`,
	}, Map{"h1": "hi", "h2": "hello"}, `<h1>hi</h1><h1>hello</h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>{title}</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component>hello</Component>`,
	}, Map{"title": "hi"}, `<h1></h1>`)
	equalMap(t, map[string]string{
		"component.duo": `<h1>{title}</h1>`,
		"main.duo":      `<script>import Component from "./component.duo";</script><Component title={h1}>hello</Component>`,
	}, Map{"title": "hi"}, `<h1></h1>`)
}

func TestSlot(t *testing.T) {
	equalMap(t, map[string]string{
		"Box.duo":  `<div class="box"><slot /></div>`,
		"main.duo": `<script>import Box from './Box.duo';</script><Box><h2>Hello!</h2><p>This is a box. It can contain anything.</p></Box>`,
	}, Map{}, `<div class="box"><h2>Hello!</h2><p>This is a box. It can contain anything.</p></div>`)
	equalMap(t, map[string]string{
		"Box.duo":  `<div class="box"></div>`,
		"main.duo": `<script>import Box from './Box.duo';</script><Box><h2>Hello!</h2><p>This is a box. It can contain anything.</p></Box>`,
	}, Map{}, `<div class="box"></div>`)
	equalMap(t, map[string]string{
		"Box.duo":  `<div class="box"><slot /></div>`,
		"main.duo": `<script>import Box from './Box.duo';</script><Box></Box>`,
	}, Map{}, `<div class="box"></div>`)
	equalMap(t, map[string]string{
		"Box.duo":  `<div class="box"><slot /></div>`,
		"main.duo": `<script>import Box from './Box.duo';</script><Box>{title}</Box>`,
	}, Map{"title": "hi"}, `<div class="box">hi</div>`)
}
