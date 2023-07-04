# .duo

Template language for Bud. Still early days.

## Features

1. Reactive frontend like Svelte
2. Supports SSR without needing to evaluate server-side JS
3. Deep reactivity using ES6 proxies (e.g. `arr.push` should trigger a re-render)
4. Client-side is built on top of [Preact](https://preactjs.com/) for a tiny footprint while supporting the React ecosystem.
5. Scoped CSS and [Tailwind](https://tailwindcss.com/) support built-in
6. Supports [Turbo Frames](https://turbo.hotwired.dev/handbook/frames)
7. Template streaming support

## License

MIT
