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

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS env variables")
	}

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
	if clientID == "" {
		clientID = "your-google-client-id.apps.googleusercontent.com"
	}
	clientSecret := os.Getenv("GOOGLE_SECRET")
	if clientSecret == "" {
		clientSecret = "your-google-client-secret"
	}

	goth.UseProviders(
		google.New(clientID, clientSecret, "http://localhost:8080/auth/google/callback", "email", "profile"),
	)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fs := http.FileServer(http.Dir("."))
	r.Handle("/css/*", fs)
	r.Handle("/js/*", fs)
	
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "robots.txt")
	})
	r.Get("/sitemap.xml", generateSitemap)

	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	})
	r.Get("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "admin/login.html")
	})
	r.Get("/admin/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "admin/dashboard.html")
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/places", getPlaces)
		r.Get("/places/{id}", getPlace)
		r.Get("/user", getUser)
		r.Get("/content", getSiteContent)
		r.Post("/contact", handleContact)
	})

	r.Get("/auth/{provider}/login", authLogin)
	r.Get("/auth/{provider}/callback", authCallback)
	r.Get("/auth/{provider}/logout", authLogout)

	r.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", adminLogin)
		r.Post("/change-credentials", adminChangeCredentials)
		r.Post("/content", updateSiteContent)
		r.Post("/places", createPlace)
		r.Put("/places/{id}", updatePlace)
		r.Delete("/places/{id}", deletePlace)
		r.Get("/messages", getMessagesAdmin)
		r.Post("/messages/{id}/reply", replyMessageAdmin)
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}