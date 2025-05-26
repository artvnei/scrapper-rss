package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
	"rss-scraper/db"
	"rss-scraper/models"
)

var (
	templates = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/login.html",
		"templates/feeds.html",
		"templates/news.html",
	))
	sessions = make(map[string]bool)
	sessMu   sync.Mutex

	adminUser = getEnv("ADMIN_USER", "admin")
	adminPass = getEnv("ADMIN_PASS", "password")
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func StartAdminServer() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/toggle", authMiddleware(toggleHandler))
	http.HandleFunc("/feeds", authMiddleware(feedsHandler))
	http.HandleFunc("/news", authMiddleware(newsHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
	})

	log.Print("[admin] listening on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("admin server error: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == adminUser && password == adminPass {
			id := uuid.NewString()
			sessMu.Lock()
			sessions[id] = true
			sessMu.Unlock()
			http.SetCookie(w, &http.Cookie{Name: "session", Value: id, Path: "/", HttpOnly: true})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
	templates.ExecuteTemplate(w, "login.html", nil)
}

func isAuthenticated(r *http.Request) bool {
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}
	sessMu.Lock()
	defer sessMu.Unlock()
	return sessions[c.Value]
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func feedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds, _ := db.GetAllRss()
	data := struct {
		Feeds []models.RssFread
	}{feeds}
	templates.ExecuteTemplate(w, "layout", data)
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	news, _ := db.GetRecentNews(20)
	counts, _ := db.CountNewsByStatus()
	data := struct {
		News   []models.News
		Counts map[string]int64
	}{news, counts}
	templates.ExecuteTemplate(w, "layout", data)
}

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	idStr := r.FormValue("id")
	activeStr := r.FormValue("active")
	id, err := uuid.Parse(idStr)
	if err == nil {
		db.SetRssActive(id, activeStr == "true")
	}
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}
