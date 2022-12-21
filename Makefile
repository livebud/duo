GREP ?=

test:
	@ ./node_modules/.bin/pegjs -o jsparser/jsparser.js jsparser/jsparser.pegjs
	@ ./node_modules/.bin/mocha --grep=$(GREP) jsparser/test.js

.PHONY: test test.partial

# Adjust the test based on
test.partial: GREP="attribute-multiple"
test.partial: test

test.watch:
	@ watch --clear -- $(MAKE) test.partial
