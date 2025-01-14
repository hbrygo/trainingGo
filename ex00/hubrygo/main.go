package main

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

type NameData struct {
	FirstName string
	LastName  string
	ImagePath string
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
		if error != nil {
			fmt.Printf("1\n")
			os.Exit(1)
		}
	}
	fileInfo, error := file.Stat()
	if error != nil {
		fmt.Printf("2\n")
		os.Exit(1)
	}
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

func parseImage(w http.ResponseWriter, r *http.Request) (string, error) {
	file, handler, err := r.FormFile("profilePic")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		fmt.Println("Error retrieving file:", err)
		return "", err
	}
	defer file.Close()

	savePath := "image/" + handler.Filename
	fmt.Printf("savePath: %s\n", savePath)
	outFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		fmt.Println("Error saving file:", err)
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		fmt.Println("Error writing file:", err)
		return "", err
	}
	return handler.Filename, nil
}

func encodeImage(newFile *os.File, newImage image.Image, format string) error {
	ext := strings.ToLower(format)

	switch ext {
	case "jpeg", "jpg":
		return jpeg.Encode(newFile, newImage, nil)
	case "png":
		return png.Encode(newFile, newImage)
	case "gif":
		return gif.Encode(newFile, newImage, nil)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
}

func grayscale(savePath string) (string, error) {
	file, err := os.Open("image/" + savePath)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return "", err
	}
	defer file.Close()

	Image, format, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %s\n", err)
		return "", err
	}

	newImage := image.NewGray(Image.Bounds())
	for y := Image.Bounds().Min.Y; y < Image.Bounds().Max.Y; y++ {
		for x := Image.Bounds().Min.X; x < Image.Bounds().Max.X; x++ {
			oldColor := Image.At(x, y)
			r, g, b, _ := oldColor.RGBA()
			gray := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256)
			newImage.SetGray(x, y, color.Gray{Y: gray})
		}
	}

	newPath := "gray_" + savePath
	newFile, err := os.Create("image/" + newPath)
	if err != nil {
		fmt.Printf("Error creating file: %s\n", err)
		return "", err
	}
	defer newFile.Close()

	err = encodeImage(newFile, newImage, format)
	if err != nil {
		fmt.Printf("Error encoding image: %s\n", err)
		return "", err
	}

	return newPath, nil
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

	savePath, err := parseImage(w, r)
	if err != nil {
		return
	}
	newPath, err := grayscale(savePath)
	if err != nil {
		return
	}
	savePath = "http://localhost:8080/image/" + newPath
	data := NameData{
		FirstName: firstName[0],
		LastName:  lastName[0],
		ImagePath: savePath,
	}
	pageName := r.URL.Path[1:]
	tmpl, err := template.ParseFiles(pageName)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
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
