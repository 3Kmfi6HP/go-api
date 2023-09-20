package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"
	// "os"
	"github.com/gorilla/mux"
	"github.com/pretty66/websocketproxy"
	"golang.org/x/net/websocket"
	"log"
	"os/exec"
	"strings"
)

func main() {
	router := mux.NewRouter()
	wp, err := websocketproxy.NewProxy("ws://127.0.0.1:7861/ws", func(r *http.Request) error {
		// Permission to verify
		// r.Header.Set("Cookie", "----")
		// Source of disguise
		// r.Header.Set("Origin", "http://82.157.123.54:9010")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	// proxy path
	http.HandleFunc("/ws", wp.Proxy)
	// router.HandleFunc("/ws", proxyWebSocket)
	router.HandleFunc("/upload", uploadFile).Methods("POST")
	router.HandleFunc("/download", downloadFile).Methods("GET")
	router.HandleFunc("/bash", executeBashCommand).Methods("POST")
	router.HandleFunc("/cat", catFile).Methods("POST")
	router.HandleFunc("/", handle).Methods("GET")
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("static")),
		),
	)
	http.Handle("/", router)
	fmt.Println("API server is running on port 7860")
	http.ListenAndServe(":7860", nil)
}
func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
func handle(w http.ResponseWriter, r *http.Request) {
	// You might want to move ParseGlob outside of handle so it doesn't
	// re-parse on every http request.
	tmpl, err := template.ParseGlob("templates/*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := ""
	if r.URL.Path == "/" {
		name = "index.html"
	} else {
		name = path.Base(r.URL.Path)
	}

	data := struct {
		Time time.Time
	}{
		Time: time.Now(),
	}

	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("error", err)
	}
}
func proxyWebSocket(w http.ResponseWriter, r *http.Request) {
	target := "wss://edtunnel.pages.dev"
	proxy := websocket.Handler(func(ws *websocket.Conn) {
		conn, err := websocket.Dial(target, "", "http://localhost")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		go io.Copy(conn, ws)
		io.Copy(ws, conn)
	})
	proxy.ServeHTTP(w, r)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ioutil.WriteFile("uploaded_file.txt", fileData, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully\n")
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Missing 'filename' parameter", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func executeBashCommand(w http.ResponseWriter, r *http.Request) {
	var command struct {
		Cmd string `json:"cmd"`
	}

	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	output, err := exec.Command(command.Cmd).CombinedOutput()
	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func catFile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Filename string `json:"filename"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadFile(request.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 200 {
		lines = lines[:200]
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strings.Join(lines, "\n")))
}
