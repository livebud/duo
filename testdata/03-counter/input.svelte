<script>
  let { count = 10 } = $props()
  function increment() {
    count += 1
  }
</script>

<button onclick={increment}>
  Clicked {count || 0}
  {count == 1 ? "time" : "times"}
</button>
