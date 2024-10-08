package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

const (
	MAX_UPLOAD_SIZE  = 10 * 1024 * 1024
	TOTP_SECRET_FILE = "secrets/totp_secret"
	UPLOAD_DIR       = "uploads"
)

var totpSecret string

var imagesUrl = os.Getenv("IMAGES_URL")

// UploadResponse is the structure of the JSON response for uploads
type UploadResponse struct {
	Filename string `json:"filename"`
	Message  string `json:"message"`
}

func SaveTOTPSecret(secret string) error {
	return os.WriteFile(TOTP_SECRET_FILE, []byte(secret), 0600)
}

func LoadTOTPSecret() (string, error) {
	data, err := os.ReadFile(TOTP_SECRET_FILE)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func InitializeTOTPSecret() {
	var err error
	if _, err = os.Stat(TOTP_SECRET_FILE); os.IsNotExist(err) {

		key, err := totp.Generate(totp.GenerateOpts{Issuer: "FileUploadApp", AccountName: "FileUploadApp"})
		if err != nil {
			log.Fatal("Failed to generate TOTP secret key")
		}
		totpSecret = key.Secret()
		if err := SaveTOTPSecret(totpSecret); err != nil {
			log.Fatal("Failed to save TOTP secret to file")
		}
		fmt.Println("Generated and saved new TOTP secret.")
	} else {
		// File exists, load the TOTP secret from the file
		totpSecret, err = LoadTOTPSecret()
		if err != nil {
			log.Fatal("Failed to load TOTP secret from file")
		}
		fmt.Println("Loaded TOTP secret from file.")
	}
}

// Serve the HTML page for file upload
func uploadPageHandler(w http.ResponseWriter, r *http.Request) {

	htmlContent := `
	<!DOCTYPE html>
	<!-- 1 -->
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Image Upload with TOTP</title>
 		<script src="https://cdnjs.cloudflare.com/ajax/libs/otpauth/9.3.2/otpauth.umd.min.js"></script>
		<style>
			body { background-color: #f0f0f0; padding: 20px; margin: 0; }
			* {    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, Vazirmatn, Courier New, monospace; font-feature-settings: normal; font-variation-settings: normal;}

		</style>
	</head>
	<body>
		<h1>Upload Image</h1>
		<label for="fileInput">Choose Image:</label>
		<input type="file" id="fileInput" accept="image/*">
		<div><button id="uploadButton">Upload</button></div>
		<p id="responseMessage"></p>

		<script>
			const otpAuth = localStorage.getItem("otpAuth");
			if (!otpAuth) {
				const token = prompt("Please enter your TOTP token"); if (token) { localStorage.setItem("otpAuth", token); document.location.reload();}
			}
			else {
				let totp = new OTPAuth.TOTP({algorithm: "SHA1",digits: 6,period: 30,secret: otpAuth});

				document.getElementById('uploadButton').addEventListener('click', async function() {
					const fileInput = document.getElementById('fileInput').files[0];
					const totpToken = totp.generate();
					const formData = new FormData();
					formData.append('image', fileInput);
					formData.append('totp_token', totpToken);
	
					try {
						const response = await fetch('/upload', {
							method: 'POST',
							body: formData
						});
						const result = await response.json();
						if (response.ok) {
							document.getElementById('responseMessage').innerHTML = "Upload successful: <a href='%s/"+ result.filename +"'>" + result.filename + "</a>";
						} else {
							document.getElementById('responseMessage').innerText = "Upload failed:" + result.message;
						}
					} catch (err) {
						document.getElementById('responseMessage').innerText = "Error: " + err.message;
					}
				});
			}
		</script>
	</body>
	</html>`
	fmt.Fprint(w, fmt.Sprintf(htmlContent, imagesUrl))
}

func ValidateTOTP(token string) bool {
	return totp.Validate(token, totpSecret)
}

// Handle the file upload with TOTP validation and return JSON
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	totpToken := r.FormValue("totp_token")
	if totpToken == "" || !ValidateTOTP(totpToken) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "Invalid or missing TOTP token",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)

	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "The uploaded file is too big.",
		})
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "Unable to retrieve file from form data.",
		})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "Unable to read file.",
		})
		return
	}

	fileType := http.DetectContentType(fileBytes)
	if fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/gif" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "File format is not supported. Only JPEG, PNG, and GIF images are allowed.",
		})
		return
	}

	randomFileName := uuid.New().String()

	fileExtension := filepath.Ext(handler.Filename)
	if fileExtension == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "File does not have a valid extension.",
		})
		return
	}

	if err := os.MkdirAll(UPLOAD_DIR, os.ModePerm); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "Unable to create upload directory.",
		})
		return
	}

	newFileName := randomFileName + fileExtension
	newFilePath := filepath.Join(UPLOAD_DIR, newFileName)

	if err := os.WriteFile(newFilePath, fileBytes, os.ModePerm); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UploadResponse{
			Message: "Unable to save the file.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{
		Filename: newFileName,
		Message:  "File uploaded successfully",
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	InitializeTOTPSecret()
	fmt.Println("Images URL:", imagesUrl)

	if err := os.MkdirAll(UPLOAD_DIR, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", uploadPageHandler)
	mux.HandleFunc("/upload", uploadFileHandler)

	handler := corsMiddleware(mux)

	fmt.Println("Starting server at :8086")
	if err := http.ListenAndServe(":8086", handler); err != nil {
		log.Fatal(err)
	}
}
