<script>
  let { todoList = $bindable([]) } = $props()
  let newItem = $state("")

  function addToList() {
    todoList = [...todoList, { text: newItem, status: false }]
    newItem = ""
  }

  function removeFromList(index) {
    todoList.splice(index, 1)
  }
</script>

<input bind:value={newItem} type="text" placeholder="new todo item" />
<button onclick={addToList} disabled={newItem === ""}>Add</button>

<br />
{#each todoList as item, index}
  <input bind:checked={item.status} type="checkbox" />
  <span class:checked={item.status}>{item.text}</span>
  <span onclick={() => removeFromList(index)}>❌</span>
  <br />
{/each}

<style>
  .checked {
    text-decoration: line-through;
  }
</style>
