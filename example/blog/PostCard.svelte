<script lang="typescript">
  import Time from "./Time.svelte"
  import { Post } from "./types"
  export let post: Post
</script>

<a class="card layout vertical center center-center" href={post.path}>
  <div class="last-updated">
    <Time relative={post.date} />
  </div>
  <h2 class="title">{post.h1}</h2>
  <div class="author layout center">
    <div class="avatar"></div>
    <div class="name">Matt Mueller</div>
  </div>
  <div class="preview"><slot /></div>
</a>

<style>
  .card {
    transition:
      background-color ease-in 150ms,
      outline-color ease-in 150ms;
    outline: 32px solid rgba(0, 0, 0, 0);
    border-radius: 8px;
    width: 100%;
    max-width: 90ch;
    text-align: center;
  }
  .card:hover {
    background: #1c1839;
    outline-color: #1c1839;
  }
  .card:hover .title {
    text-decoration-color: currentColor;
  }
  .last-updated {
    margin-top: 20px;
    color: #7e73bb;
    font-size: 12px;
  }
  .title {
    margin-top: 4px;
    color: #edf0ff;
    font-size: 32px;
    font-weight: 500;
    line-height: 1.1875;
    transition: text-decoration-color ease-in 150ms;
    text-decoration: underline;
    text-decoration-color: transparent;
  }
  .author {
    margin-top: 16px;
  }
  .avatar {
    border-radius: 9999px;
    min-height: 24px;
    min-width: 24px;
    max-height: 24px;
    max-width: 24px;
    background-image: url(https://www.gravatar.com/avatar/824b41e6108a22c4c96f50ee23419369?s=48);
    background-size: contain;
    background-repeat: no-repeat;
  }
  .name {
    margin-left: 8px;
    color: #9c99af;
    font-size: 14px;
  }
  .preview {
    margin-top: 16px;
    aspect-ratio: 16 / 9;
    border-radius: 16px;
    width: 100%;
    padding: 0 8px;
    overflow: hidden;
  }
  .preview :global(img) {
    width: 100%;
    border-radius: 16px;
  }
</style>
