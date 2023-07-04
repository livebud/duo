import { render as preactRender, h, hydrate } from 'https://cdn.jsdelivr.net/npm/preact@10.15.1/+esm'
import Proxy from 'https://esm.run/internal/proxy'

export function render(Component, target, props = {}) {
  const proxy = Proxy(props)
  const component = Component(h, proxy)
  hydrate(h(component, proxy, []), target)
  window.requestAnimationFrame(() => {
    props.subscribe(() => {
      preactRender(h(component, props, []), target)
    })
  })
}
