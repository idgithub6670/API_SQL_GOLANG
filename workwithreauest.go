package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Coruse struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Instructor string  `json:"instructor"`
}

var CoruseList []Coruse

func init() {
	CoruseJSON := `[
		{
			"id":101,
			"name":"Python",
			"price":2590,
			"instructor":"Bornto Dev"
		},

		{
			"id":102,
			"name":"Java",
			"price":2590999,
			"instructor":"Bornto Dev"
		},

		{
			"id":103,
			"name":"GO",
			"price":2000,
			"instructor":"Bornto Dev"
		}
	]`
	err := json.Unmarshal([]byte(CoruseJSON), &CoruseList)
	if err != nil {
		log.Fatal(err)
	}
}

func getNextID() int {
	highestID := -1
	for _, course := range CoruseList {
		if highestID < course.ID {
			highestID = course.ID

		}
	}
	return highestID + 1
}

func courseHandler(w http.ResponseWriter, r *http.Request) {
	coruseJSON, err := json.Marshal(CoruseList)
	switch r.Method {
	case http.MethodGet:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(coruseJSON)
	case http.MethodPost:
		var newCourse Coruse
		Bodybyte, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(Bodybyte, &newCourse)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if newCourse.ID != 0 {
			w.WriteHeader(http.StatusBadRequest)
			return

		}
		newCourse.ID = getNextID()
		CoruseList = append(CoruseList, newCourse)
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func main() {
	http.HandleFunc("/course", courseHandler)
	http.ListenAndServe(":5000", nil)
}
