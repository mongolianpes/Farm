package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"project-farm/internal/announcements"
	"project-farm/internal/handlers"
	"project-farm/internal/images"
	"project-farm/internal/messenger"
	"project-farm/internal/session"
)

const (
	sitePort = ":443"
)

func main() {
	hand := handlers.NewHand()
	defer hand.DB.Close()

	go session.OldSessionsRemover(hand.DB)

	mux := http.NewServeMux()

	mux.HandleFunc("/{$}", hand.HomePageHandler)
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern == "" {
			http.ServeFile(w, r, handlers.HTMLPagesPath+"404.html")
			return
		}

		mux.ServeHTTP(w, r)
	})

	server := &http.Server{
		Handler: handler,
		Addr:    sitePort,
	}

	go server.ListenAndServe()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Не удалось завершить работу обработчика запросов", "error", err)
	}

	if err := images.CloseConnectionToService(); err != nil {
		slog.Error("Не удалось разоврвать соединение с микросервисом Images", "error", err)
	}

	if err := announcements.CloseConnectionToService(); err != nil {
		slog.Error("Не удалось разоврвать соединение с микросервисом Announcements", "error", err)
	}

	if err := messenger.CloseConnectionToService(); err != nil {
		slog.Error("Не удалось разоврвать соединение с микросервисом Messenger", "error", err)
	}

	if err := hand.DB.Close(); err != nil {
		slog.Error("Не удалось разоврвать соединение с базой данных", "error", err)
	}
}
