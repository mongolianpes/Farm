package handlers

import (
	"html/template"
	"time"

	"project-farm/internal/messenger"
	"project-farm/internal/rdb"
)

type Handler struct {
	Tmpl      *template.Template
	RedisDB   rdb.DB
	Messenger messenger.Messenger
}

const timeToCompleteRequest = 30 * time.Second

const htmlPages = "./htmlPages/*.html"

func NewHand() *Handler {
	rdb, err := rdb.NewClient("rdb:6379")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseGlob(htmlPages)
	if err != nil {
		panic(err)
	}

	messenger, err := messenger.NewClient()
	if err != nil {
		panic(err)
	}

	return &Handler{
		Tmpl:      tmpl,
		RedisDB:   rdb,
		Messenger: messenger,
	}
}
