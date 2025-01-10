# ex00
## Grayscaler

## Goal

Write a webserver which lets users upload any PNG file and returns it in grayscale.

The server, when started, will serve the HTTP route `POST /grayscale` and will expect a PNG file in the
`multipart/form-data` encoding. The server then turns the PNG image grayscale and returns it in `image/png` encoding.

The server will also serve a HTML file with a form to use the

## Instructions

- The program should be written in Go.
- The program can only use Go's standard library, including the external library (`golang.org/x/*`).
- The program will be compiled with:
```sh
go build -o grayscaled ./cmd/grayscaled
```
- The program should not panic.

## Bonus
Complete as many or as little as you want.

- Handle more image formats (JPG, GIF, etc.
- Add an option to choose the color of the grayscaled image. (example: instead of white to black gradient, make it a white to red gradient).
- Instead of directly returning the modified image, cache it, serve it on its own URL and return the URL to the user.
- Anything else ... be creative!
