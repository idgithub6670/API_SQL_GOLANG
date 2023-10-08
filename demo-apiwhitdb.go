package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

const coursePath = "courses"
const BasePath = "/api"

type Course struct {
	CourseID   int     `json:"courseid"`
	CourseName string  `json:"coursename"`
	Price      float64 `json:"price"`
	ImageURL   string  `json:"imageurl"`
}

func SetupDB() {
	var err error
	Db, err = sql.Open("mysql", "root:210658Za!@tcp(127.0.0.1:3306)/coursedb")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(Db)
	Db.SetConnMaxLifetime(time.Minute * 3)
	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(10)
}

func getCourseList() ([]Course, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()
	results, err := Db.QueryContext(ctx, `SELECT
	courseid,
	coursename,
	price,
	imageurl
	From courseonline`)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer results.Close()
	courses := make([]Course, 0)
	for results.Next() {
		var course Course
		results.Scan(&course.CourseID,
			&course.CourseName,
			&course.Price,
			&course.ImageURL)
		courses = append(courses, course)
	}
	return courses, nil
}

func insertProduct(course Course) (int, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()
	result, err := Db.ExecContext(ctx, `INSERT INTO courseonline
    (courseid,
    coursename,
	price,
    imageurl) 
    VALUES (?,?,?,?)`, course.CourseID, course.CourseName, course.Price, course.ImageURL)

	if err != nil {
		log.Print(err.Error())
		return 0, err
	}
	InsertID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		return 0, err
	}
	return int(InsertID), nil
}

func getCourse(courseid int) (*Course, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := Db.QueryRowContext(ctx, `SELECT
	courseid,
	coursename,
	price,
	imageurl
	FROM courseonline
	WHERE courseid = ?`, courseid)

	course := &Course{}
	err := row.Scan(
		&course.CourseID,
		&course.CourseName,
		&course.Price,
		&course.ImageURL)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Println(err)
		return nil, err
	}
	return course, nil

}

func removeCourse(courseID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Db.ExecContext(ctx, `DELETE FROM courseonline WHERE courseid = ?`, courseID)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
func handleCourses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		courseList, err := getCourseList()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		j, err := json.Marshal(courseList)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		var course Course
		err := json.NewDecoder(r.Body).Decode(&course)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		CourseID, err := insertProduct(course)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"courseid":%d}`, CourseID)))
	case http.MethodOptions:
		return
	}

}

func handleCourse(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", coursePath))
	if len(urlPathSegments[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	courseID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		course, err := getCourse(courseID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if course == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		j, err := json.Marshal(course)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodDelete:
		err := removeCourse(courseID)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}
func corsmiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("ccess-Control-Allow-Methods", "Accept,Content-Type,Content-Leng")
		handler.ServeHTTP(w, r)
	})
}

func SetupRoutes(apiBasePath string) {
	courseHandler := http.HandlerFunc(handleCourse)
	http.Handle(fmt.Sprintf("%s/%s/", apiBasePath, coursePath), corsmiddleware(courseHandler))

	coursesHandler := http.HandlerFunc(handleCourses)
	http.Handle(fmt.Sprintf("%s/%s", apiBasePath, coursePath), corsmiddleware(coursesHandler))
}

func main() {
	SetupDB()
	SetupRoutes(BasePath)
	log.Fatal(http.ListenAndServe(":5000", nil))

}
