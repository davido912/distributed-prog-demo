package teacherportal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dist_project/app/grades"
	"github.com/dist_project/app/registry"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers() {
	http.Handle("/", http.RedirectHandler("/students", http.StatusPermanentRedirect))

	h := new(studentsHandler)

	http.Handle("/students", h)
	http.Handle("/students/", h)
}

type studentsHandler struct{}

func (sh studentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")

	switch len(pathSegments) {
	case 2:
		sh.renderStudents(w, r)
	case 3:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		sh.renderStudent(w, r, id)
	case 4:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if strings.ToLower(pathSegments[3]) != "grades" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		sh.renderGrades(w, r, id)
	default:
		w.WriteHeader(http.StatusNotFound)
	}

}

func (studentsHandler) renderStudents(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error retrieving students: ", err)

		}
	}()

	serviceURL, err := registry.GetProvider(registry.GradingService)
	if err != nil {
		return
	}

	res, err := http.Get(serviceURL + "/students")
	if err != nil {
		return
	}

	var s grades.Students
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&s)
	if err != nil {
		return
	}

	rootTemplate.Lookup("students.gohtml").Execute(w, s)
}

func (studentsHandler) renderStudent(w http.ResponseWriter, r *http.Request, id int) {
	var err error
	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error retrieving students: ", err)

		}
	}()

	serviceURL, err := registry.GetProvider(registry.GradingService)
	if err != nil {
		return
	}

	res, err := http.Get(fmt.Sprintf("%v/students/%v", serviceURL, id))
	if err != nil {
		return
	}

	var s grades.Student
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&s)
	if err != nil {
		return
	}

	rootTemplate.Lookup("student.gohtml").Execute(w, s)
}

func (studentsHandler) renderGrades(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		w.Header().Add("location", fmt.Sprintf("/students/%v", id))
		w.WriteHeader(http.StatusTemporaryRedirect)
	}()

	title := r.FormValue("Title")
	_type := r.FormValue("Type")
	score, err := strconv.ParseFloat(r.FormValue("Score"), 32)
	if err != nil {
		log.Println("Failed to parse score: ", err)
		return
	}
	g := grades.Grade{
		Title: title,
		Type:  grades.GradeType(_type),
		Score: float32(score),
	}

	data, err := json.Marshal(g)
	if err != nil {
		log.Println("failed to convert grade to json: ", g, err)
		return
	}

	serviceURL, err := registry.GetProvider(registry.GradingService)
	if err != nil {
		log.Println("Failed to retrieve instance of Grading Service ", err)
		return
	}

	res, err := http.Post(fmt.Sprintf("%v/students/%v/grades", serviceURL, id),
		"application/json", bytes.NewBuffer(data))

	if err != nil {
		log.Println("Failed to save grade to grading service", err)
		return
	}

	if res.StatusCode != http.StatusCreated {
		log.Println("Failed to save grade to Grading Services. Status: ", res.StatusCode)
		return
	}

}
