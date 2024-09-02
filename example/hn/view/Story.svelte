<script>
  // import { format as timeago } from "timeago.js"
  export let story = {}
  function formatURL(url) {
    if (!url) return ""
    const parsed = new URL(url)
    return parsed.host
  }

  function formatComments(num_comments) {
    switch (num_comments) {
      case 1:
        return "1 comment"
      default:
        return `${num_comments || 0} comments`
    }
  }
</script>

<div class="story">
  <div>
    <a class="title" href={story.URL || `/${story.ID}`}>{story.Title}</a>
    {#if story.URL}
      <a class="url" href={story.URL}>({story.URL})</a>
    {/if}
  </div>
  <div class="meta">
    {story.Points} points by {story.Author} • {story.CreatedAt} •
    <a href={`/${story.ID}`}>{story.NumComments} comments</a>
  </div>
</div>

<style>
  .story {
    padding: 10px;
    font-size: 14px;
  }
  a {
    text-decoration: none;
    color: inherit;
  }
  a[href]:hover {
    text-decoration: underline;
  }
  .title {
    font-weight: 500;
  }
  .url,
  .meta {
    color: gray;
  }
</style>
