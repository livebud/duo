<script>
  export let greeting = "hello"
  setInterval(() => {
    greeting += "o"
  }, 1000)
</script>

<h1 name="hello-{greeting}-cool">{greeting}</h1>
