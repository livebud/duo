package evaluator_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/livebud/duo/internal/evaluator"
	"github.com/livebud/duo/internal/parser"
	"github.com/matthewmueller/diff"
)

func equal(t *testing.T, name, input string, props interface{}, expected string) {
	t.Helper()
	if name == "" {
		name = input
	}
	t.Run(name, func(t *testing.T) {
		t.Helper()
		doc, err := parser.Parse(input)
		if err != nil {
			t.Fatal(err)
		}
		evaluator := evaluator.New(doc)
		str := new(strings.Builder)
		actual := ""
		if err := evaluator.Evaluate(str, props); err != nil {
			actual = err.Error()
		} else {
			// TODO: remove this, this should happen earlier
			actual = str.String()
			// actual = strings.TrimSpace(str.String())
			// actual = strings.ReplaceAll(actual, "  ", "")
			// actual = strings.ReplaceAll(actual, "\t", "")
			// actual = strings.ReplaceAll(actual, "\n", "")
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
