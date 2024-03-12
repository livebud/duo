<script lang="typescript">
  import Header from "./Header.svelte"
  import Footer from "./Footer.svelte"
  import Time from "./Time.svelte"
  import type { Post } from "./types"
  export let post: Post
</script>

<!-- <svelte:head>
  <title>{title}</title>
  <meta
    name="viewport"
    content="width=device-width, initial-scale=1, shrink-to-fit=no"
  />
  <meta name="robots" content="index, follow" />
  <meta
    name="googlebot"
    content="index, follow, max-snippet:-1, max-image-preview:large, max-video-preview:-1"
  />
  <meta
    name="bingbot"
    content="index, follow, max-snippet:-1, max-image-preview:large, max-video-preview:-1"
  />
  <meta property="og:locale" content={props.locale} />
  <meta property="og:site_name" content={props.site_name} />
  <meta name="author" content="Matt Mueller" />
  <meta name="description" content={props.description} />
  <meta property="og:url" content={canonical} />
  <meta property="og:title" content={title} />
  <link rel="canonical" href={canonical} />
  <meta property="og:description" content={props.description} />
  {@html images}
  <meta property="og:type" content="article" />
  <meta name="article:published_time" content={timestamp} />
  <meta property="og:author" content={author} />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:site" content={props.twitter.handle} />
  <meta name="twitter:creator" content={props.twitter.handle} />
</svelte:head> -->

<div class="post">
  <Header home={true} intro={false} />
  <div class="mountains"></div>
  <div class="heading layout vertical center">
    <h1>{post.h1}</h1>
    <div class="meta layout center">
      <div class="last-updated">
        Last Updated <Time relative={post.date} />
      </div>
      <div class="author layout center">
        <div class="avatar"></div>
        <div class="name">Matt Mueller</div>
      </div>
    </div>
  </div>
  <article class="layout vertical center">
    <div class="markdown"><slot /></div>
  </article>
  <Footer />
</div>
