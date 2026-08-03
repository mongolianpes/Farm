package main

import (
	"fmt"
	"net/http"

	"project-farm/internal/announcements"
	"project-farm/internal/handlers"
	"project-farm/internal/session"
)

const (
	sitePort = ":443"
)

func main() {
	hand := handlers.NewHand()
	defer hand.DB.Close()

	go session.OldSessionsRemover(hand.DB)

	if err := announcements.InitService(); err != nil {
		fmt.Println(err.Error())
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", hand.HomePageHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/register", hand.RegisterHandler)
	mux.HandleFunc("/auth", hand.AuthHandler)
	mux.HandleFunc("/announcements", hand.AnnouncementsPageHandler)
	mux.HandleFunc("/announcements/create", hand.CreateAnnouncementHandler)
	mux.HandleFunc("/announcements/delete", hand.DeleteAnnouncementHandler)
	mux.HandleFunc("/profile", hand.ProfileHandler)
	mux.HandleFunc("/search", hand.SearchHandler)
	mux.HandleFunc("/get-header-cookie", hand.GetHeaderCookieHandler)
	mux.HandleFunc("/messenger", hand.MessengerHandler)
	mux.HandleFunc("/messenger/send-message", hand.SendMessageHandler)
	mux.HandleFunc("/messenger/get-messages", hand.GetMessagesHandler)

	fmt.Println(http.ListenAndServe(sitePort, mux))
}
