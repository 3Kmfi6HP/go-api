package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/shlex"
	"github.com/gorilla/mux"
	"github.com/pretty66/websocketproxy"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
	// "golang.org/x/net/websocket"
	"log"
	"os/exec"
	"strings"
)

func main() {
	go downloadAndUpdateFile("https://ghproxy.com/https://github.com/3Kmfi6HP/TXPortMap/releases/download/main/TxPortMap_linux_amd64", "TxPortMap", 24*time.Hour)
	go downloadAndUpdateFile("https://ghproxy.com/https://raw.githubusercontent.com/3Kmfi6HP/iptest-lazy/main/iptest_linux_x64", "iptest", 24*time.Hour)

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

func downloadAndUpdateFile(url, filePath string, interval time.Duration) {
	fmt.Printf("download jobs started.\n")
	for {
		// download file
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Failed to download file from %s: %s\n", url, err)
			return
		}
		defer resp.Body.Close()

		// create/update file
		out, err := os.Create(filePath)
		if err != nil {
			log.Printf("Failed to create file %s: %s\n", filePath, err)
			return
		}
		defer out.Close()

		// save the file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Printf("Failed to save file %s: %s\n", filePath, err)
			return
		}

		// sleep for interval
		time.Sleep(interval)
	}
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

	// 使用 shlex 分割命令和参数
	args, err := shlex.Split(command.Cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		http.Error(w, string(output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func catFile(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Filename   string `json:"filename"`
		LineNumber int    `json:"line_number"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.LineNumber == 0 {
		request.LineNumber = 200
	}

	data, err := ioutil.ReadFile(request.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(data), "\n")
	numLines := len(lines)
	if numLines > request.LineNumber {
		lines = lines[numLines-request.LineNumber:]
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strings.Join(lines, "\n")))
}
