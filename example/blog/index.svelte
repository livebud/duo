<script lang="typescript">
  import Header from "./Header.svelte"
  import PostCard from "./PostCard.svelte"
  import Footer from "./Footer.svelte"
  import type { Post } from "./types"
  export let posts: Post[] = []
</script>

<!-- <svelte:head>
  <title>Matt Mueller</title>
  <link rel="icon" href="/static/favicon.ico" type="image/x-icon" />
</svelte:head> -->

<div class="home">
  <Header home={false} intro={true} />
  <div class="mountains"></div>
  <div class="timeline layout vertical center">
    <div class="heading">TIMELINE</div>
    {#each posts as post}
      <div class="post layout center vertical">
        <div class="line"></div>
        <PostCard {post} />
      </div>
    {/each}
  </div>
  <Footer />
</div>

<style>
  .home {
    background: linear-gradient(
        1deg,
        rgba(34, 40, 72, 0.2) 2.3%,
        rgba(0, 0, 0, 0) 126.24%
      ),
      #131023;
    min-height: 100vh;
  }
  .mountains {
    background-image: url(/static/mountains.svg);
    background-size: contain;
    background-repeat: repeat-x;
    background-position: right top;
    width: 100%;
    height: 404px;
  }
  .timeline {
    margin-top: -300px;
  }
  .heading {
    border-radius: 26px;
    border: 2px solid #211f43;
    background: #110f1f;
    color: #a2b1ff;
    font-size: 14px;
    line-height: 24px; /* 171.429% */
    letter-spacing: 2.1px;
    text-align: center;
    padding: 8px 36px;
  }
  .line {
    width: 2px;
    height: 82px;
    background: linear-gradient(180deg, #120f23 0%, #201e42 100%);
  }
  .post {
    width: 100%;
  }
</style>
