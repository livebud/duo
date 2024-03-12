<script lang="typescript">
  import Link from "./Link.svelte"
  export let home = false
  export let intro = false
</script>

<header class="layout vertical center">
  {#if home}
    <Link href="/"><div class="left-stars"></div></Link>
  {:else}
    <div class="left-stars"></div>
  {/if}
  <div class="right-stars"></div>
  <ul class="menu layout">
    {#if home}
      <li>
        <Link href="/">Home</Link>
      </li>
    {:else}
      <li>
        <Link href="https://twitter.com/mattmueller">Twitter</Link>
      </li>
      <li>
        <Link href="https://github.com/matthewmueller">GitHub</Link>
      </li>
    {/if}
  </ul>
  {#if intro}
    <div class="about layout vertical center">
      <div class="avatar"></div>
      <div class="intro layout vertical center">
        <h1 class="hello">Hey! I&apos;m Matt</h1>
        <p class="description">
          I&apos;m a full-stack developer engineer at{" "}
          <a href="https://stripe.com">Stripe</a>. In my free-time, I hack on
          <a href="https://standupjack.com">side projects</a>
          and{" "}
          <a href="https://github.com/matthewmueller">
            contribute to open source
          </a>
          . I enjoy traveling slowly, running and lifting.
        </p>
      </div>
    </div>
  {/if}
</header>

<style>
  header {
    position: relative;
    max-width: 100%;
    min-height: 333px;
    overflow-x: hidden;
    padding: 0 8px 0 8px;
  }
  .menu {
    margin: 0 auto;
    padding: 36px;
    z-index: 1;
    color: #7e73bb;
    font-size: 14px;
  }
  .menu li + li {
    margin-left: 32px;
  }
  .menu li {
    transition: color ease-in 175ms;
  }
  .menu li:hover {
    color: #edf0ff;
  }
  .left-stars {
    position: absolute;
    top: 0;
    left: 0;
    background: url(/static/left-stars.svg) no-repeat;
    background-size: contain;
    background-position: top left;
    width: 100%;
    max-width: 90px;
    height: 50px;
    aspect-ratio: 1.8;
  }
  .right-stars {
    position: absolute;
    top: 0;
    right: 0;
    background: url(/static/right-stars.svg) no-repeat;
    background-size: contain;
    background-position: top right;
    width: 100%;
    max-width: 139px;
    height: 90px;
    aspect-ratio: 1.54;
  }
  @media (min-width: 730px) {
    .left-stars {
      max-width: 182px;
      height: 101px;
    }
    .right-stars {
      max-width: 231px;
      height: 150px;
    }
  }
  @media (min-width: 910px) {
    .left-stars {
      max-width: 272px;
      height: 151px;
    }
    .right-stars {
      max-width: 308px;
      height: 200px;
    }
  }
  @media (min-width: 1440px) {
    .left-stars {
      max-width: 540px;
      height: 300px;
    }
    .right-stars {
      max-width: 513px;
      height: 333px;
    }
  }
  .about {
    padding: 64px 0;
    z-index: 1;
  }
  @media (min-width: 1080px) {
    .about {
      flex-direction: row;
    }
    .intro {
      align-items: start;
    }
  }
  .avatar {
    border-radius: 9999px;
    min-height: 148px;
    min-width: 148px;
    max-height: 148px;
    max-width: 148px;
    background-image: url(https://www.gravatar.com/avatar/824b41e6108a22c4c96f50ee23419369?s=296);
    background-size: contain;
    background-repeat: no-repeat;
  }
  .intro {
    padding: 32px;
  }
  .hello {
    font-weight: 500;
    color: #a2b1ff;
    font-size: 28px;
  }
  .description {
    max-width: 48ch;
    color: #edf0ff;
    font-size: 20px;
    font-style: normal;
    font-weight: 400;
    line-height: 30px; /* 150% */
  }

  .description a {
    color: #a2b1ff;
    text-decoration: underline;
    text-decoration-color: transparent;
    transition: text-decoration-color ease-in 50ms;
  }

  .description a[href]:hover {
    text-decoration-color: currentColor;
  }
</style>
