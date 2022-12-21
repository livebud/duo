### action

```svelte
<input use:autofocus>

```

### action-duplicate

```svelte
<input use:autofocus use:autofocus>
```

### action-with-call

```svelte
<input use:tooltip="{t('tooltip msg')}">

```

### action-with-identifier

```svelte
<input use:tooltip={message}>

```

### action-with-literal

```svelte
<input use:tooltip="{'tooltip msg'}">

```

### animation

```svelte
{#each things as thing (thing)}
	<div animate:flip>flips</div>
{/each}
```

### attribute-class-directive

```svelte
<div class:foo={isFoo}></div>
```

### attribute-containing-solidus

```svelte
<a href=https://www.google.com>Google</a>

```

### attribute-curly-bracket

```svelte
<input foo=a{1} />

```

### attribute-dynamic

```svelte
<div style='color: {color};'>{color}</div>

```

### attribute-dynamic-boolean

```svelte
<textarea readonly={readonly}></textarea>

```

### attribute-empty

```svelte
<div a="" b={''} c='' d="{''}" ></div>
```

### attribute-empty-error

```svelte
<div class= ></div>
```

### attribute-escaped

```svelte
<div data-foo='semi:&quot;space:&quot letter:&quote number:&quot1 end:&quot'></div>

```

### attribute-multiple

```svelte
<div id='x' class='y'></div>
```

### attribute-shorthand

```svelte
<div {id}/>
```

### attribute-static

```svelte
<div class='foo'></div>
```

### attribute-static-boolean

```svelte
<textarea readonly></textarea>
```

### attribute-style

```svelte
<div style="color: red;">red</div>
```

### attribute-style-directive

```svelte
<div style:color={myColor}></div>
```

### attribute-style-directive-modifiers

```svelte
<div style:color|important={myColor}></div>
```

### attribute-style-directive-shorthand

```svelte
<div style:color></div>
```

### attribute-style-directive-string

```svelte
<div style:color="red"></div>
<div style:color='red'></div>
<div style:color=red></div>
<div style:color="red{variable}"></div>
<div style:color='red{variable}'></div>
<div style:color=red{variable}></div>
<div style:color={`template${literal}`}></div>
```

### attribute-unique-binding-error

```svelte
<Widget foo={42} bind:foo/>
```

### attribute-unique-error

```svelte
<div class='foo' class='bar'></div>
```

### attribute-unique-shorthand-error

```svelte
<div title='foo' {title}></div>
```

### attribute-unquoted

```svelte
<div class=foo></div>
```

### attribute-with-whitespace

```svelte
<button on:click= {foo}>Click</button>

```

### await-catch

```svelte
{#await thePromise}
	<p>loading...</p>
{:catch theError}
	<p>oh no! {theError.message}</p>
{/await}
```

### await-then-catch

```svelte
{#await thePromise}
	<p>loading...</p>
{:then theValue}
	<p>the value is {theValue}</p>
{:catch theError}
	<p>oh no! {theError.message}</p>
{/await}
```

### binding

```svelte
<script>
	let name;
</script>

<input bind:value={name}>
```

### binding-shorthand

```svelte
<script>
	let foo;
</script>

<Widget bind:foo/>
```

### comment

```svelte
<!-- a comment -->
```

### comment-with-ignores

```svelte
<!-- svelte-ignore foo bar -->
```

### component-dynamic

```svelte
<svelte:component this="{foo ? Foo : Bar}"></svelte:component>
```

### convert-entities

```svelte
Hello &amp; World
```

### convert-entities-in-element

```svelte
<p>Hello &amp; World</p>
```

### css

```svelte
<div>foo</div>

<style>
	div {
		color: red;
	}
</style>
```

### css-option-none

```svelte
<div>foo</div>

<style>
	div {
		color: red;
	}
</style>

```

### dynamic-element-string

```svelte
<svelte:element this="div"></svelte:element>
<svelte:element this="div" class="foo"></svelte:element>
```

### dynamic-element-variable

```svelte
<svelte:element this={tag}></svelte:element>
<svelte:element this={tag} class="foo"></svelte:element>
```

### dynamic-import

```svelte
<script>
	import { onMount } from 'svelte';

	onMount(() => {
		import('./foo.js').then(foo => {
			console.log(foo.default);
		});
	});
</script>
```

### each-block

```svelte
{#each animals as animal}
	<p>{animal}</p>
{/each}

```

### each-block-destructured

```svelte
<script>
	export let animals;
</script>

{#each animals as [key, value, ...rest]}
	<p>{key}: {value}</p>
{/each}

```

### each-block-else

```svelte
{#each animals as animal}
	<p>{animal}</p>
{:else}
	<p>no animals</p>
{/each}

```

### each-block-indexed

```svelte
{#each animals as animal, i}
	<p>{i}: {animal}</p>
{/each}

```

### each-block-keyed

```svelte
{#each todos as todo (todo.id)}
	<p>{todo}</p>
{/each}

```

### element-with-attribute

```svelte
<span attr="foo"></span>
<span attr='bar'></span>

```

### element-with-attribute-empty-string

```svelte
<span attr=""></span>
<span attr=''></span>

```

### element-with-mustache

```svelte
<h1>hello {name}!</h1>

```

### element-with-text

```svelte
<span>test</span>
```

### elements

```svelte
<!doctype html>
```

### error-catch-before-closing

```svelte
{#await true}
	{#each foo as bar}
{:catch f}
{/await}
```

### error-catch-without-await

```svelte
{:catch theValue}

```

### error-comment-unclosed

```svelte
<!-- an unclosed comment
```

### error-css

```svelte
<style>
	this is not css
</style>
```

### error-css-global-without-selector

```svelte
<style>
	:global {}
</style>

```

### error-else-before-closing

```svelte
{#if true}
	<li>
{:else}
{/if}
```

### error-else-before-closing-2

```svelte
{#if true}
	{#await p}
{:else}
{/if}
```

### error-else-before-closing-3

```svelte
<li>
{:else}
```

### error-else-if-before-closing

```svelte
{#if true}
	{#await foo}
{:else if false}
{/if}
```

### error-else-if-before-closing-2

```svelte
{#if true}
	<p>
{:else if false}
{/if}
```

### error-else-if-without-if

```svelte
{#await foo}
{:then bar}
	{:else if}
{/await}
```

### error-empty-attribute-shorthand

```svelte
<span {}></span>
```

### error-empty-classname-binding

```svelte
<h1 class:={true}>Hello</h1>

```

### error-empty-directive-name

```svelte
<h1 use:>Hello</h1>

```

### error-illegal-expression

```svelte
{42 = nope}

```

### error-multiple-styles

```svelte
<div>foo</div>

<style>
	div {
		color: red;
	}
</style>

<style>
	div {
		color: blue;
	}
</style>
```

### error-script-unclosed

```svelte
<h1>Hello {name}!</h1>

<script>
```

### error-script-unclosed-eof

```svelte
<script>

<h1>Hello {name}!</h1>
```

### error-self-reference

```svelte
<svelte:self/>
```

### error-style-unclosed

```svelte
<h1>Hello {name}!</h1>

<style>
```

### error-style-unclosed-eof

```svelte
<style>

<h1>Hello {name}!</h1>
```

### error-svelte-selfdestructive

```svelte
{#if x}
	<svelte:selfdestructive x="{x - 1}"/>
{/if}
```

### error-then-before-closing

```svelte
{#await true}
	<li>
{:then f}
{/await}
```

### error-then-without-await

```svelte
{:then theValue}

```

### error-unclosed-attribute-self-close-tag

```svelte
<Component test={{a: 1} />
```

### error-unexpected-end-of-input

```svelte
<div>
```

### error-unexpected-end-of-input-b

```svelte
<d
```

### error-unexpected-end-of-input-c

```svelte
<
```

### error-unexpected-end-of-input-d

```svelte
{#if foo}
	<p>foo</p>
```

### error-unmatched-closing-tag

```svelte
</div>
```

### error-unmatched-closing-tag-autoclose

```svelte
<p>
	<pre>pre tag</pre>
</p>
```

### error-unmatched-closing-tag-autoclose-2

```svelte
<div>
	<p>
	<pre>pre tag</pre>
</div>
</p>
```

### error-void-closing

```svelte
<input>this is illegal!</input>
```

### error-window-children

```svelte
<svelte:window>contents</svelte:window>
```

### error-window-duplicate

```svelte
<svelte:window/>
<svelte:window/>
```

### error-window-inside-block

```svelte
{#if foo}
	<svelte:window/>
{/if}
```

### error-window-inside-element

```svelte
<div>
	<svelte:window/>
</div>
```

### event-handler

```svelte
<button on:click="{() => visible = !visible}">toggle</button>

{#if visible}
	<p>hello!</p>
{/if}

```

### if-block

```svelte
{#if foo}bar{/if}

```

### if-block-else

```svelte
{#if foo}
	<p>foo</p>
{:else}
	<p>not foo</p>
{/if}

```

### if-block-elseif

```svelte
{#if x > 10}
	<p>x is greater than 10</p>
{:else if x < 5}
	<p>x is less than 5</p>
{/if}

```

### implicitly-closed-li

```svelte
<ul>
	<li>a
	<li>b
	<li>c
</ul>
```

### implicitly-closed-li-block

```svelte
<ul>
	<li>a
	{#if true}
		<li>b
	{/if}
	<li>c
</ul>

```

### nbsp

```svelte
<span>&nbsp;</span>
```

### no-error-if-before-closing

```svelte
{#if true}
	<input>
{:else}
{/if}

{#if true}
	<br>
{:else}
{/if}

{#await true}
	<input>
{:then f}
{/await}

{#await true}
	<br>
{:then f}
{/await}
```

### raw-mustaches

```svelte
<p> {@html raw1} {@html raw2} </p>

```

### raw-mustaches-whitespace-error

```svelte
{@htmlfoo}

```

### refs

```svelte
<script>
	let foo;
</script>

<canvas bind:this={foo}></canvas>
```

### script

```svelte
<script>
	let name = 'world';
</script>

<h1>Hello {name}!</h1>
```

### script-comment-only

```svelte
<script>
	// TODO write some code
</script>

<div></div>
```

### script-comment-trailing

```svelte
<script>
	let name = 'world';

	// trailing line comment
</script>

<h1>Hello {name}!</h1>

```

### script-comment-trailing-multiline

```svelte
<script>
	let name = 'world';

	/*
		trailing multiline comment
	*/
</script>

<h1>Hello {name}!</h1>

```

### self-closing-element

```svelte
<div/>
```

### self-reference

```svelte
{#if depth > 1}
	<svelte:self depth='{depth - 1}'/>
{/if}
```

### slotted-element

```svelte
<Component><div slot='foo'></div></Component>

```

### space-between-mustaches

```svelte
<p> {a} {b} : {c} : </p>

```

### spread

```svelte
<div {...props}></div>

```

### style-inside-head

```svelte
<svelte:head><style></style></svelte:head>
```

### textarea-children

```svelte
<textarea>
	<p>not actually an element. {foo}</p>
</textarea>
```

### textarea-end-tag

```svelte
<textarea>
	<p>not actu </textar ally an element. {foo}</p>
</textare


> </textaread >asdf</textarea


</textarea

>


```

### transition-intro

```svelte
<div in:style="{{opacity: 0}}">fades in</div>
```

### transition-intro-no-params

```svelte
<div in:fade>fades in</div>
```

### unusual-identifier

```svelte
{#each things as 𐊧}
	<p>{𐊧}</p>
{/each}
```

### whitespace-after-script-tag

```svelte
<script>
	let name = 'world';
</script     




>

<h1>Hello {name}!</h1>
```

### whitespace-after-style-tag

```svelte
<div>foo</div>

<style>
	div {
		color: red;
	}
</style     




>
```

### whitespace-leading-trailing

```svelte


				<p>just chillin' over here</p>
```

### whitespace-normal

```svelte
<h1>Hello <strong>{name}! </strong><span>How are you?</span></h1>

```

