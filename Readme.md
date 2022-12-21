# .duo

Experimental template language for Bud. Very early stages.

## Project Goals

1. Support SSR without requiring server-side JS
2. Support type-safe properties from the server.
3. Reactive on the frontend like Svelte
4. Supports tree-shaking client builds for production.

## Why?

To fully support server-side rendering a reactive frontend language like Svelte, you'd need a Javascript interpreter on the server.

Unfortunately there's not currently a single good solution for rendering JS in Go. Each solution has tradeoffs.

Instead, I'd like to experiment with a server-side rendering algorithm that will skip over any unknown expressions on the server-side and let client-side hydration fill in any gaps.

## Development

```sh
npm install
make test

# Run a specific test
GREP="attribute-multiple" make test
```

## Current Plan

- [x] Setup Svelte HTML parser in pegjs
- [ ] Get the important Svelte parser tests passing (9/110 tests currently passing)
- [ ] Translate the generated parser to Go
- [ ] Render the parsed AST into HTML

I plan on working on this here and there. Contributions very welcome!

## License

MIT
