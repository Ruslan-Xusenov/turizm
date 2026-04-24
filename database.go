package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./turizm.db")
	if err != nil {
		log.Fatal(err)
	}

	createTables()
	seedData()
}

func createTables() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE,
			name TEXT,
			avatar TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS places (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			location TEXT,
			description TEXT,
			price INTEGER,
			category TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS place_images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			place_id INTEGER,
			image_url TEXT,
			FOREIGN KEY(place_id) REFERENCES places(id)
		);`,
		`CREATE TABLE IF NOT EXISTS site_content (
			key TEXT PRIMARY KEY,
			value TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE,
			password_hash TEXT,
			requires_password_change BOOLEAN DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			email TEXT,
			message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_replied BOOLEAN DEFAULT 0
		);`,
	}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			log.Fatalf("Error creating table: %v\nQuery: %s", err, table)
		}
	}
}

func seedData() {
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&adminCount)
	if adminCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		db.Exec("INSERT INTO admins (username, password_hash, requires_password_change) VALUES (?, ?, ?)", "admin", string(hash), true)
	}

	var contentCount int
	db.QueryRow("SELECT COUNT(*) FROM site_content").Scan(&contentCount)
	if contentCount == 0 {
		defaultContent := map[string]string{
			"hero_title": "O'zbekistonning go'zalligini kashf eting",
			"hero_subtitle": "Tarixiy shaharlar, betakror tabiat va mehmondo'st odamlar sizni kutmoqda",
			"hero_img1": "https://images.unsplash.com/photo-1564507592333-c60657eea523?w=1920",
			"hero_img2": "https://images.unsplash.com/photo-1548013146-72479768bada?w=1920",
			"hero_img3": "https://images.unsplash.com/photo-1476514525535-07fb3b4ae5f1?w=1920",
			"about_title": "Nega aynan bizni tanlashingiz kerak?",
			"about_subtitle": "O'zbekiston turizmi bilan tanishtirishda yetakchi kompaniya",
			"about_text": "Biz O'zbekistonning eng go'zal joylarini butun dunyo bo'ylab odamlarga tanishtirishda 15 yildan ortiq tajribaga egamiz. Professional gidlarimiz, qulay transport xizmatlarimiz va maxsus tur paketlarimiz bilan sizga unutilmas sayohat tajribasini taqdim etamiz.",
			"contact_address": "Toshkent shahri, Chilonzor tumani",
			"contact_phone": "+998 90 123 45 67",
			"contact_email": "info@uztourism.uz",
			"destinations_title": "Mashhur Yo'nalishlar",
			"destinations_subtitle": "O'zbekistonning eng mashhur turistik shaharlarini kashf eting",
			"dest1_name": "Samarqand",
			"dest1_desc": "Registon maydoni, Go'ri Amiq va boshqa tarixiy obidalari bilan mashhur.",
			"dest1_img": "https://images.unsplash.com/photo-1564507592333-c60657eea523?w=600",
			"dest2_name": "Buxoro",
			"dest2_desc": "Islom sivilizatsiyasining markazi, 2500 yillik tarixga ega.",
			"dest2_img": "https://images.unsplash.com/photo-1548013146-72479768bada?w=600",
			"dest3_name": "Xiva",
			"dest3_desc": "Ichan-Qal'a - UNESCO tomonidan tan olingan dunyoviy meros.",
			"dest3_img": "https://images.unsplash.com/photo-1599576935803-91efed66198f?w=600",
			"dest4_name": "Toshkent",
			"dest4_desc": "Zamonaviy arxitektura va an'anaviy madaniyat uyg'unligi.",
			"dest4_img": "https://images.unsplash.com/photo-1605548230624-8d2d639e7022?w=600",
		}
		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("INSERT INTO site_content (key, value) VALUES (?, ?)")
		for k, v := range defaultContent {
			stmt.Exec(k, v)
		}
		tx.Commit()
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM places").Scan(&count)
	if err != nil || count > 0 {
		return
	}

	tx, _ := db.Begin()
	stmtPlace, _ := tx.Prepare("INSERT INTO places (name, location, description, price, category) VALUES (?, ?, ?, ?, ?)")
	stmtImage, _ := tx.Prepare("INSERT INTO place_images (place_id, image_url) VALUES (?, ?)")

	places := []struct {
		name        string
		location    string
		description string
		price       int
		category    string
		images      []string
	}{
		{
			"Registon maydoni", "Samarqand",
			"Samarqandning eng mashhur diqqatga sazovor joyi. Uch madrasa - Ulug'bek, Sherdor va Tillakori madrasalaridan iborat.",
			50000, "Arxitektura",
			[]string{"https://images.unsplash.com/photo-1564507592333-c60657eea523?w=600", "https://images.unsplash.com/photo-1476514525535-07fb3b4ae5f1?w=600"},
		},
		{
			"Go'ri Amir", "Samarqand",
			"Amir Temurning maqbarasi. Zardushtiylar dinining ta'siri ostida qurilgan bu inshoot katta sharqiy xazinalar bilan mashhur.",
			45000, "Tarixiy obida",
			[]string{"https://images.unsplash.com/photo-1583417319070-4a69db38a482?w=600", "https://images.unsplash.com/photo-1548013146-72479768bada?w=600"},
		},
		{
			"Ichan-Qal'a", "Xiva",
			"UNESCO tomonidan dunyoviy meros sifatida tan olingan Xivaning ichki shaharchasi. 2500 yillik tarixga ega.",
			60000, "Qadimiy shahar",
			[]string{"https://images.unsplash.com/photo-1599576935803-91efed66198f?w=600", "https://images.unsplash.com/photo-1605548230624-8d2d639e7022?w=600"},
		},
	}

	for _, p := range places {
		res, _ := stmtPlace.Exec(p.name, p.location, p.description, p.price, p.category)
		placeID, _ := res.LastInsertId()
		for _, img := range p.images {
			stmtImage.Exec(placeID, img)
		}
	}
	tx.Commit()
}