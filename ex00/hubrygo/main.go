package main

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
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

	// Sauvegarde le fichier
	savePath := "image/" + handler.Filename
	fmt.Printf("savePath: %s\n", savePath)
	outFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		fmt.Println("Error saving file:", err)
		return "", err
	}
	defer outFile.Close()

	// Copie le contenu du fichier téléchargé
	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		fmt.Println("Error writing file:", err)
		return "", err
	}
	return handler.Filename, nil
}

func grayscale(savePath string) (string, error) {
	// Ouverture du fichier
	file, err := os.Open("image/" + savePath)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return "", err
	}
	defer file.Close()
	Image, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %s\n", err)
		return "", err
	}
	// Création d'une nouvelle image
	newImage := image.NewRGBA(Image.Bounds())
	// Parcours de l'image
	for y := Image.Bounds().Min.Y; y < Image.Bounds().Max.Y; y++ {
		for x := Image.Bounds().Min.X; x < Image.Bounds().Max.X; x++ {
			// Récupération de la couleur du pixel
			oldColor := Image.At(x, y)
			r, g, b, a := oldColor.RGBA()
			// Calcul de la nouvelle couleur
			newColor := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
			// Affectation de la nouvelle couleur au pixel
			newImage.Set(x, y, newColor)
		}
	}
	// Sauvegarde de la nouvelle image
	newPath := "gray_" + savePath
	newFile, err := os.Create("image/" + newPath)
	if err != nil {
		fmt.Printf("Error creating file: %s\n", err)
		return "", err
	}
	defer newFile.Close()
	err = png.Encode(newFile, newImage)
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

	// file, handler, err := r.FormFile("profilePic")
	// if err != nil {
		// 	http.Error(w, "Error retrieving file", http.StatusBadRequest)
		// 	fmt.Println("Error retrieving file:", err)
		// 	return
		// }
		// defer file.Close()
		
		// // Sauvegarde le fichier
		// savePath := "image/" + handler.Filename
		// fmt.Printf("savePath: %s\n", savePath)
		// outFile, err := os.Create(savePath)
		// if err != nil {
			// 	http.Error(w, "Error saving file", http.StatusInternalServerError)
			// 	fmt.Println("Error saving file:", err)
			// 	return
			// }
			// defer outFile.Close()
			
	// // Copie le contenu du fichier téléchargé
	// _, err = io.Copy(outFile, file)
	// if err != nil {
		// 	http.Error(w, "Error writing file", http.StatusInternalServerError)
		// 	fmt.Println("Error writing file:", err)
		// 	return
		// }
		
		savePath, err := parseImage(w, r)
		if err != nil {
			return
		}
		newPath, err := grayscale(savePath)
		savePath = "http://localhost:8080/image/" + newPath
		fmt.Printf("savePath: %s\n", savePath)
		data := NameData{
			FirstName: firstName[0],
			LastName:  lastName[0],
			ImagePath: savePath,
		}
		fmt.Printf("data: %v\n", data)
		pageName := r.URL.Path[1:]
		pageName = "template/" + pageName + ".html"
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
