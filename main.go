package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Coaster struct {
	Name         string
	Manufacturer string
	ID           string
	InPark       string
	Height       int
}

type coastersHandler struct {
	sync.Mutex
	store map[string]Coaster
}

func (handler *coastersHandler) coasters(w http.ResponseWriter, r *http.Request) {
	fmt.Println("The method invoked is ", r.Method)
	switch r.Method {
	case "GET":
		handler.get(w, r)
		return

	case "POST":
		handler.post(w, r)
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (handler *coastersHandler) get(w http.ResponseWriter, r *http.Request) {
	coasters := make([]Coaster, len(handler.store))
	handler.Lock()
	i := 0
	for _, coaster := range handler.store {
		coasters[i] = coaster
		i++
	}
	handler.Unlock()
	jsonByte, err := json.Marshal(coasters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	fmt.Println("gping to send the response of rget now")
	w.Header().Add("content-type", "application/json") //optional
	w.WriteHeader(http.StatusOK)                       //optional. by default its 200
	w.Write(jsonByte)
}

func (handler *coastersHandler) post(w http.ResponseWriter, r *http.Request) {
	fmt.Println("entered POST")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")

	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content type application/json but got %s", ct)))
		return
	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	fmt.Println("the coaster is  ", coaster)

	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	handler.Lock()
	handler.store[coaster.ID] = coaster
	defer handler.Unlock()
}

func (handler *coastersHandler) getRandomCoaster(w http.ResponseWriter, r *http.Request) {
	ids := make([]string, len(handler.store))
	handler.Lock()
	i := 0
	for id := range handler.store {
		ids[i] = id
		i++
	}

	defer handler.Unlock()

	var target string
	if len(ids) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 {
		target = ids[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))]
	}

	fmt.Println(target)
}

func (handler *coastersHandler) getCoaster(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if parts[2] == "random" {
		handler.getRandomCoaster(w, r)
	}
	// coasters := make([]Coaster, len(handler.store))
	handler.Lock()
	coaster, ok := handler.store[parts[2]]
	handler.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	jsonByte, err := json.Marshal(coaster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	fmt.Println("gping to send the response of rget now")
	w.Header().Add("content-type", "application/json") //optional
	w.WriteHeader(http.StatusOK)                       //optional. by default its 200
	w.Write(jsonByte)
}

func newCoasterHandler() *coastersHandler {
	return &coastersHandler{
		store: map[string]Coaster{},
	}
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		panic("requreied env password not set")
	}
	return &adminPortal{password: password}
}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()

	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 unauthorized"))
		return
	}

	w.Write([]byte("<html><h1>Super secret admin portal</h1></html>"))
}

//2nd arg is default arg
func main() {
	admin := newAdminPortal()
	fmt.Println("the coaster is  ")
	coastersHandlers := newCoasterHandler()
	http.HandleFunc("/coasters", coastersHandlers.coasters)
	http.HandleFunc("/coasters/", coastersHandlers.getCoaster)
	http.HandleFunc("/admin", admin.handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
