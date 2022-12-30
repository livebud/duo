GREP ?= "($(shell cat jsparser/test.txt | sed '/^\#/d' | tr '\n' '|' | sed 's/|$$//'))$$"

test.all:
	@ ./node_modules/.bin/pegjs -o jsparser/jsparser.js jsparser/jsparser.pegjs
	@ ./node_modules/.bin/mocha jsparser/test.js

test:
	@ ./node_modules/.bin/pegjs -o jsparser/jsparser.js jsparser/jsparser.pegjs
	@ ./node_modules/.bin/mocha --grep=$(GREP) jsparser/test.js

.PHONY: test test.partial

test.watch:
	@ watch --clear -- $(MAKE) test

lex:
	go test ./internal/lexer

lex.watch:
	@ watch --clear -- $(MAKE) lex
