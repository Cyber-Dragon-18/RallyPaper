package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Team struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Stop1       int    `json:"stop1"`
	Stop2       int    `json:"stop2"`
	Stop3       int    `json:"stop3"`
	Stop4       int    `json:"stop4"`
	Stop5       int    `json:"stop5"`
	Stop6       int    `json:"stop6"`
	TotalPoints int    `json:"total_points"`
}

type Stage struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Clue     string `json:"clue"`
	MaxPts   int    `json:"max_points"`
	StopNum  int    `json:"stop_number"`
}

func main() {
	var err error

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Println("WARNING: DATABASE_URL not set")
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTables()

	mux := http.NewServeMux()

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./public"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	// Teams
	mux.HandleFunc("/api/teams", handleTeams)
	mux.HandleFunc("/api/teams/score", handleUpdateScore) // must be before /api/teams/
	mux.HandleFunc("/api/teams/", handleTeamByID)

	// Stages
	mux.HandleFunc("/api/stages", handleStages)
	mux.HandleFunc("/api/stages/", handleStageByID)

	// Leaderboard
	mux.HandleFunc("/api/leaderboard", handleLeaderboard)

	log.Println("Rally Paper running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func createTables() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS teams (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT DEFAULT '#1a73e8',
			stop1 INTEGER DEFAULT 0,
			stop2 INTEGER DEFAULT 0,
			stop3 INTEGER DEFAULT 0,
			stop4 INTEGER DEFAULT 0,
			stop5 INTEGER DEFAULT 0,
			stop6 INTEGER DEFAULT 0,
			total_points INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS stages (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			location TEXT,
			clue TEXT,
			max_points INTEGER DEFAULT 26,
			stop_number INTEGER UNIQUE NOT NULL
		);`,
	}
	for _, t := range tables {
		if _, err := db.Exec(t); err != nil {
			log.Println("Error creating table:", err)
		}
	}
}

// ---------- TEAMS ----------

func handleTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query(`SELECT id, name, color, stop1, stop2, stop3, stop4, stop5, stop6, total_points FROM teams ORDER BY id`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var teams []Team
		for rows.Next() {
			var t Team
			rows.Scan(&t.ID, &t.Name, &t.Color, &t.Stop1, &t.Stop2, &t.Stop3, &t.Stop4, &t.Stop5, &t.Stop6, &t.TotalPoints)
			teams = append(teams, t)
		}
		if teams == nil {
			teams = []Team{}
		}
		json.NewEncoder(w).Encode(teams)

	case http.MethodPost:
		var t Team
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if t.Color == "" {
			t.Color = "#1a73e8"
		}
		err := db.QueryRow(
			`INSERT INTO teams (name, color) VALUES ($1, $2) RETURNING id`,
			t.Name, t.Color,
		).Scan(&t.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handleTeamByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := strings.TrimPrefix(r.URL.Path, "/api/teams/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		db.Exec(`DELETE FROM teams WHERE id = $1`, id)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPut:
		var t Team
		json.NewDecoder(r.Body).Decode(&t)
		db.Exec(`UPDATE teams SET name=$1, color=$2 WHERE id=$3`, t.Name, t.Color, id)
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handleUpdateScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var payload struct {
		TeamID  int `json:"team_id"`
		StopNum int `json:"stop_number"`
		Points  int `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if payload.StopNum < 1 || payload.StopNum > 6 {
		http.Error(w, "stop_number must be 1-6", 400)
		return
	}

	col := "stop" + strconv.Itoa(payload.StopNum)

	_, err := db.Exec(
		`UPDATE teams SET `+col+` = $1,
			total_points = stop1+stop2+stop3+stop4+stop5+stop6
		WHERE id = $2`,
		payload.Points, payload.TeamID,
	)
	if err != nil {
		log.Println("Score update error:", err)
		http.Error(w, err.Error(), 500)
		return
	}

	// Recalculate total (the inline expression doesn't include the new value yet)
	db.Exec(`UPDATE teams SET total_points = stop1+stop2+stop3+stop4+stop5+stop6 WHERE id = $1`, payload.TeamID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ---------- STAGES ----------

func handleStages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query(`SELECT id, name, location, clue, max_points, stop_number FROM stages ORDER BY stop_number`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var stages []Stage
		for rows.Next() {
			var s Stage
			rows.Scan(&s.ID, &s.Name, &s.Location, &s.Clue, &s.MaxPts, &s.StopNum)
			stages = append(stages, s)
		}
		if stages == nil {
			stages = []Stage{}
		}
		json.NewEncoder(w).Encode(stages)

	case http.MethodPost:
		var s Stage
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if s.MaxPts == 0 {
			s.MaxPts = 26
		}
		err := db.QueryRow(
			`INSERT INTO stages (name, location, clue, max_points, stop_number) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
			s.Name, s.Location, s.Clue, s.MaxPts, s.StopNum,
		).Scan(&s.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

func handleStageByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := strings.TrimPrefix(r.URL.Path, "/api/stages/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", 400)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var s Stage
		json.NewDecoder(r.Body).Decode(&s)
		db.Exec(`UPDATE stages SET name=$1, location=$2, clue=$3, max_points=$4, stop_number=$5 WHERE id=$6`,
			s.Name, s.Location, s.Clue, s.MaxPts, s.StopNum, id)
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		db.Exec(`DELETE FROM stages WHERE id = $1`, id)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}

// ---------- LEADERBOARD ----------

func handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rows, err := db.Query(`SELECT id, name, color, stop1, stop2, stop3, stop4, stop5, stop6, total_points FROM teams ORDER BY total_points DESC`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var t Team
		rows.Scan(&t.ID, &t.Name, &t.Color, &t.Stop1, &t.Stop2, &t.Stop3, &t.Stop4, &t.Stop5, &t.Stop6, &t.TotalPoints)
		teams = append(teams, t)
	}
	if teams == nil {
		teams = []Team{}
	}
	json.NewEncoder(w).Encode(teams)
}
