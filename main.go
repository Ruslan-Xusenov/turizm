package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var router *chi.Mux

func init() {
	godotenv.Load()
	initDB()

	key := os.Getenv("SESSION_SECRET")
	if key == "" {
		key = "secret_session_key"
	}
	maxAge := 86400 * 30
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	gothic.Store = store

	clientID := os.Getenv("GOOGLE_KEY")
	clientSecret := os.Getenv("GOOGLE_SECRET")
	
	callbackURL := os.Getenv("CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = "http://localhost:8080/auth/google/callback"
	}

	if clientID != "" && clientSecret != "" {
		goth.UseProviders(
			google.New(clientID, clientSecret, callbackURL, "email", "profile"),
		)
	}

	router = chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	fs := http.FileServer(http.Dir("."))
	router.Handle("/css/*", fs)
	router.Handle("/js/*", fs)
	
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "robots.txt")
	})
	router.Get("/sitemap.xml", generateSitemap)

	router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	})
	router.Get("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "admin/login.html")
	})
	router.Get("/admin/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "admin/dashboard.html")
	})

	router.Route("/api", func(r chi.Router) {
		r.Get("/places", getPlaces)
		r.Get("/places/{id}", getPlace)
		r.Get("/user", getUser)
		r.Get("/content", getSiteContent)
		r.Post("/contact", handleContact)
	})

	router.Get("/auth/{provider}/login", authLogin)
	router.Get("/auth/{provider}/callback", authCallback)
	router.Get("/auth/{provider}/logout", authLogout)

	router.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", adminLogin)
		r.Post("/change-credentials", adminChangeCredentials)
		r.Post("/content", updateSiteContent)
		r.Post("/places", createPlace)
		r.Put("/places/{id}", updatePlace)
		r.Delete("/places/{id}", deletePlace)
		r.Get("/messages", getMessagesAdmin)
		r.Post("/messages/{id}/reply", replyMessageAdmin)
	})
}

// Handler for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}