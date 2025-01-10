package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
)

type NameData struct {
	FirstName []string
	LastName  []string
	imagePath []string
}

func getMethod(w http.ResponseWriter, r *http.Request) {
	pageName := r.URL.Path[1:]
	if pageName == "" {
		pageName = "home"
		fmt.Printf("nothing\n")
	}
	pageName = "template/" + pageName + ".html"
	file, error := os.Open(pageName)
	if error != nil {
		fmt.Printf("404 not found\n")
		file, error = os.Open("template/404.html")
	}
	fileInfo, error := file.Stat()
	if error != nil {
		fmt.Printf("2\n")
		os.Exit(1)
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func postMethod(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // Limite à 10 Mo
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		fmt.Println("Error parsing form:", err)
		return
	}

	firstName := r.Form["firstName"]
	lastName := r.Form["lastName"]

	file, handler, err := r.FormFile("profilePic")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		fmt.Println("Error retrieving file:", err)
		return
	}
	defer file.Close()

	// Sauvegarde le fichier
	savePath := "image/" + handler.Filename
	outFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		fmt.Println("Error saving file:", err)
		return
	}
	defer outFile.Close()

	// Copie le contenu du fichier téléchargé
	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		fmt.Println("Error writing file:", err)
		return
	}

	savePath = "http://localhost:8080/" + handler.Filename
	data := NameData{
		FirstName: firstName,
		LastName:  lastName,
		imagePath: []string{savePath},
	}
	fmt.Printf("data: %v\n", data)
	pageName := r.URL.Path[1:]
	pageName = "template/" + pageName + ".html"
	tmpl, err := template.ParseFiles(pageName)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	// imagePath := "<img src='" + savePath + "' alt='gopher' style='width:235px;height:320px;'>"
	tmpl.Execute(w, data)
}

func getPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Printf("got Get request\n")
		getMethod(w, r)
	case "POST":
		fmt.Printf("got Post request\n")
		postMethod(w, r)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {
	http.HandleFunc("/", getPage)
	http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("image"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	} else if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else {
		fmt.Printf("Server started\n")
	}
}
