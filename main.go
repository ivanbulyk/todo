package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type List struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Position    int    `json:"position"`
}

type Comment struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Date string `json:"date"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:postgres@localhost/todo")
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

// CREATE DATABASE todo;
// \c todo;
// CREATE TABLE projects (
// 	id SERIAL NOT NULL PRIMARY KEY,
// 	name VARCHAR(500) NOT NULL,
// 	description  VARCHAR(1000)
// 	);
// Add some initial values...
// INSERT INTO projects (name, description)
// VALUES('Create TODO1', 'Create TODO basic structure1'),
// VALUES('Create TODO2', 'Create TODO basic structure2'),
// VALUES('Create TODO3', 'Create TODO basic structure3'),
// VALUES('Create TODO4', 'Create TODO basic structure4');
// Your driver info:
// db, err := sql.Open("postgres", "postgres://user:password@localhost/databasename")
// In project root run:
// go run main.go
func main() {
	http.HandleFunc("/projects", projectsIndex)
	http.HandleFunc("/projects/show", projectsShow)
	http.HandleFunc("/projects/create", projectsCreate)
	http.HandleFunc("/projects/delete", projectsDelete)
	http.HandleFunc("/projects/update", projectsUpdate)
	http.ListenAndServe(":3000", nil)
}

// Fetching all projects
// curl -i localhost:3000/projects
func projectsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	rows, err := db.Query("SELECT * FROM projects ORDER BY ID ASC")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	prjs := make([]*Project, 0)
	for rows.Next() {
		pr := new(Project)
		err := rows.Scan(&pr.ID, &pr.Description, &pr.Name)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		prjs = append(prjs, pr)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	out, err := json.Marshal(prjs)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, string(out))

}

// Fetching specific project
// curl -i "localhost:3000/projects/show?ID=1
func projectsShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	ID, _ := strconv.Atoi(r.FormValue("ID"))
	if ID == 0 {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	row := db.QueryRow("SELECT * FROM projects  WHERE ID = $1", ID)

	pr := new(Project)
	err := row.Scan(&pr.ID, &pr.Name, &pr.Description)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	out, err := json.Marshal(pr)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Fprintf(w, string(out))
}

// Creating project
// curl -i -X POST -d "ID=6&Name=Create TODO6&Description=Create TODO basic structure6" localhost:3000/projects/create
func projectsCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	ID, err := strconv.ParseInt(r.FormValue("ID"), 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	Name := r.FormValue("Name")
	Description := r.FormValue("Description")
	if Name == "" || Description == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	result, err := db.Exec("INSERT INTO projects VALUES($1, $2, $3)", ID, Name, Description)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	out, err := json.Marshal(Name)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	fmt.Fprintf(w, "Project %v created successfully (%v row affected)\n", string(out), rowsAffected)
}

// Deleting project
// curl  -X DELETE  localhost:3000/projects/delete?ID=3
func projectsDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	ID, err := strconv.ParseInt(r.FormValue("ID"), 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	result, err := db.Exec("DELETE FROM projects WHERE ID = $1", ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	out, err := json.Marshal(ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	fmt.Fprintf(w, "Project %v deleted successfully (%v row affected)\n", string(out), rowsAffected)

}

// Updating project
// curl -X PUT -d 'ID=6&Name=Create TODO667&Description=Create TODO basic structure667' localhost:3000/projects/update
func projectsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	ID, err := strconv.ParseInt(r.FormValue("ID"), 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	Name := r.FormValue("Name")
	Description := r.FormValue("Description")
	if Name == "" || Description == "" {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	result, err := db.Exec("UPDATE projects	SET Name = $2, Description = $3	WHERE ID = $1", ID, Name, Description)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	out, err := json.Marshal(ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	fmt.Fprintf(w, "Project %v updated successfully, (%v row affected)\n", string(out), rowsAffected)
}
