<script lang="typescript">
  // Based on: https://github.com/component/relative-date/blob/master/index.js

  const second = 1000
  const minute = 60 * second
  const hour = 60 * minute
  const day = 24 * hour
  const week = 7 * day
  const year = day * 365
  const month = year / 12

  function timeAgo(date: number, now?: number) {
    now = now || Date.now()
    const ms = Math.abs(now - date)

    if (ms < second) return ""

    if (ms == second) return "one second"
    if (ms < minute) return Math.ceil(ms / second) + " seconds"

    if (ms == minute) return "one minute"
    if (ms < hour) return Math.ceil(ms / minute) + " minutes"

    if (ms == hour) return "one hour"
    if (ms < day) return Math.ceil(ms / hour) + " hours"

    if (ms == day) return "one day"
    if (ms < week) return Math.ceil(ms / day) + " days"

    if (ms == week) return "one week"
    if (ms < month) return Math.ceil(ms / week) + " weeks"

    if (ms == month) return "one month"
    if (ms < year) return Math.ceil(ms / month) + " months"

    if (ms == year) return "one year"
    return Math.round(ms / year) + " years"
  }

  export let relative: number = Date.now()
</script>

<time dateTime={new Date(relative).toUTCString()}>
  {timeAgo(relative)} ago
</time>
