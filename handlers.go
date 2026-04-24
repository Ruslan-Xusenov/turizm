package handler

import (
	"encoding/json"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
)

type Place struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Price       int      `json:"price"`
	Category    string   `json:"category"`
	Images      []string `json:"images"`
}

func getPlaces(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, location, description, price, category FROM places")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var places []Place
	for rows.Next() {
		var p Place
		if err := rows.Scan(&p.ID, &p.Name, &p.Location, &p.Description, &p.Price, &p.Category); err != nil {
			continue
		}

		imgRows, _ := db.Query("SELECT image_url FROM place_images WHERE place_id = ?", p.ID)
		for imgRows.Next() {
			var img string
			imgRows.Scan(&img)
			p.Images = append(p.Images, img)
		}
		imgRows.Close()

		places = append(places, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func getPlace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p Place
	err := db.QueryRow("SELECT id, name, location, description, price, category FROM places WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.Location, &p.Description, &p.Price, &p.Category)
	if err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}

	imgRows, _ := db.Query("SELECT image_url FROM place_images WHERE place_id = ?", p.ID)
	for imgRows.Next() {
		var img string
		imgRows.Scan(&img)
		p.Images = append(p.Images, img)
	}
	imgRows.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func authLogin(w http.ResponseWriter, r *http.Request) {
	if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
		saveUser(gothUser.UserID, gothUser.Email, gothUser.Name, gothUser.AvatarURL)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

func generateSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>https://uztourism.uz/</loc>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
    <url>
        <loc>https://uztourism.uz/admin/login</loc>
        <changefreq>monthly</changefreq>
        <priority>0.5</priority>
    </url>
</urlset>`))
}

func authCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	saveUser(user.UserID, user.Email, user.Name, user.AvatarURL)

	session, _ := gothic.Store.Get(r, "app_session")
	session.Values["user_id"] = user.UserID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func authLogout(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)

	session, _ := gothic.Store.Get(r, "app_session")
	session.Values["user_id"] = ""
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func saveUser(id, email, name, avatar string) {
	db.Exec("INSERT INTO users (id, email, name, avatar) VALUES (?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET name=excluded.name, avatar=excluded.avatar",
		id, email, name, avatar)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	session, _ := gothic.Store.Get(r, "app_session")
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var name, email string
	db.QueryRow("SELECT name, email FROM users WHERE id = ?", userID).Scan(&name, &email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_in", "name": name, "email": email})
}

// ==============================================
// CMS and ADMIN HANDLERS
// ==============================================

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangeCredentialsPayload struct {
	Username    string `json:"username"`
	OldPassword string `json:"old_password"`
	NewUsername string `json:"new_username"`
	NewPassword string `json:"new_password"`
}

func adminLogin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var storedHash string
	var requiresChange bool
	err := db.QueryRow("SELECT password_hash, requires_password_change FROM admins WHERE username = ?", creds.Username).Scan(&storedHash, &requiresChange)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if requiresChange {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "requires_password_change",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": creds.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"token":  tokenString,
	})
}

func adminChangeCredentials(w http.ResponseWriter, r *http.Request) {
	var payload ChangeCredentialsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM admins WHERE username = ?", payload.Username).Scan(&storedHash)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(payload.OldPassword)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	
	_, err = db.Exec("UPDATE admins SET username = ?, password_hash = ?, requires_password_change = 0 WHERE username = ?", payload.NewUsername, string(newHash), payload.Username)
	if err != nil {
		http.Error(w, "Error updating credentials", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getSiteContent(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT key, value FROM site_content")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	content := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		content[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

func updateSiteContent(w http.ResponseWriter, r *http.Request) {
	// Simple auth check via header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if token == nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("INSERT INTO site_content (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value")
	for k, v := range payload {
		stmt.Exec(k, v)
	}
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func validateAdmin(r *http.Request) bool {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return false
	}
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	return token != nil && token.Valid
}

func createPlace(w http.ResponseWriter, r *http.Request) {
	if !validateAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var p Place
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	tx, _ := db.Begin()
	res, err := tx.Exec("INSERT INTO places (name, location, description, price, category) VALUES (?, ?, ?, ?, ?)",
		p.Name, p.Location, p.Description, p.Price, p.Category)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	placeID, _ := res.LastInsertId()
	stmt, _ := tx.Prepare("INSERT INTO place_images (place_id, image_url) VALUES (?, ?)")
	for _, img := range p.Images {
		if img != "" {
			stmt.Exec(placeID, img)
		}
	}
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "id": placeID})
}

func updatePlace(w http.ResponseWriter, r *http.Request) {
	if !validateAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	var p Place
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	tx, _ := db.Begin()
	_, err := tx.Exec("UPDATE places SET name=?, location=?, description=?, price=?, category=? WHERE id=?",
		p.Name, p.Location, p.Description, p.Price, p.Category, id)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Exec("DELETE FROM place_images WHERE place_id=?", id)
	stmt, _ := tx.Prepare("INSERT INTO place_images (place_id, image_url) VALUES (?, ?)")
	for _, img := range p.Images {
		if img != "" {
			stmt.Exec(id, img)
		}
	}
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deletePlace(w http.ResponseWriter, r *http.Request) {
	if !validateAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	tx, _ := db.Begin()
	tx.Exec("DELETE FROM place_images WHERE place_id=?", id)
	tx.Exec("DELETE FROM places WHERE id=?", id)
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

type ContactMessage struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	IsReplied bool   `json:"is_replied"`
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&msg)

	_, err := db.Exec("INSERT INTO messages (name, email, message) VALUES (?, ?, ?)", msg.Name, msg.Email, msg.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getMessagesAdmin(w http.ResponseWriter, r *http.Request) {
	if !validateAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, _ := db.Query("SELECT id, name, email, message, created_at, is_replied FROM messages ORDER BY created_at DESC")
	defer rows.Close()

	var msgs []ContactMessage
	for rows.Next() {
		var m ContactMessage
		rows.Scan(&m.ID, &m.Name, &m.Email, &m.Message, &m.CreatedAt, &m.IsReplied)
		msgs = append(msgs, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

func replyMessageAdmin(w http.ResponseWriter, r *http.Request) {
	if !validateAdmin(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	var req struct {
		ReplyText string `json:"reply_text"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var email string
	db.QueryRow("SELECT email FROM messages WHERE id = ?", id).Scan(&email)

	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	if smtpUser != "" && smtpPass != "" {
		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		msg := []byte("To: " + email + "\r\n" +
			"Subject: UzTourism - Xabaringizga javob\r\n" +
			"\r\n" +
			req.ReplyText + "\r\n")
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{email}, msg)
		if err != nil {
			http.Error(w, "Email yuborishda xatolik: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		println("No SMTP credentials provided in .env. Faking email to:", email, "with text:", req.ReplyText)
	}

	db.Exec("UPDATE messages SET is_replied = 1 WHERE id = ?", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
