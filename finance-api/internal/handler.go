package internal

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type user struct {
	id    int
	name  string
	email string
	roles string
}

type UserDto struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

func GetCurrentUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Fetching current user")
		// Query the database to get the list of people
		rows, err := db.Query("SELECT id, name, email, roles FROM users LIMIT 1")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Iterate over the rows and create a slice of Person structs
		var u user
		rows.Next()
		if err := rows.Scan(&u.id, &u.name, &u.email, &u.roles); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse the JSON data into a map[string]interface{}
		var rolesJson map[string]interface{}
		err = json.Unmarshal([]byte(u.roles), &rolesJson)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var roles []string
		for k := range rolesJson {
			roles = append(roles, k)
		}

		// Marshal the slice of Person structs to JSON
		jsonData, err := json.Marshal(UserDto{
			ID:    u.id,
			Name:  u.name,
			Email: u.email,
			Roles: roles,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header and write the JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}
