package controller

import (
	"net/http"
	"strconv"

	"github.com/livebud/duo"
	"github.com/matthewmueller/hackernews"
)

func New(hn *hackernews.Client, view *duo.View) *Controller {
	return &Controller{hn, view}
}

type Controller struct {
	hn   *hackernews.Client
	view *duo.View
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	stories, err := c.hn.FrontPage(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.view.Render(w, "index.svelte", map[string]any{
		"stories": stories,
	})
}

func (c *Controller) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	story, err := c.hn.Find(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.view.Render(w, "show.svelte", map[string]any{
		"story": story,
	})
}
