package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"html/template"
	"io"
	"log"
	//"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"upload_images/models"
)

type ViewData struct {
	Title       string
	Description string
}

type Image struct {
	models.Image
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func GetUploadForm(w http.ResponseWriter, r *http.Request) {
	data := ViewData{
		Title:       "Uploading file",
		Description: "Put your file",
	}
	t, _ := template.ParseFiles("templates/upload.html")

	t.Execute(w, data)
}

func UploadHandler (w http.ResponseWriter, r *http.Request) {
	maxSize := int64(32 << 20)

	err := r.ParseMultipartForm(maxSize)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprintf(w, "Image too large. Max Size: %v", maxSize)
		return
	}

	files := r.MultipartForm.File["images"]

	fmt.Println(files)

	for image, _ := range files {
		file, err := files[image].Open()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		extension := filepath.Ext(files[image].Filename)

		randomName := randomFilename(extension)

		defer file.Close()

		out, err := os.Create("./images/" + randomName)

		defer out.Close()

		if err != nil {
			fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
			return
		}

		_, err = io.Copy(out, file)

		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		s, err := session.NewSession(&aws.Config{
			Region:                            aws.String(os.Getenv("S3_REGION")),
			Credentials:                       credentials.NewStaticCredentials(
				os.Getenv("AWS_SECRET_ID") ,
				os.Getenv("AWS_SECRET_KEY") ,
				"",
			),
		})

		if err != nil {
			fmt.Fprintf(w, "Could not connectc to AWS")
		}

		fileName, err := UploadFileToS3(s, randomName)

		if err != nil {
			fmt.Fprintf(w, "Could not upload file : ")
			fmt.Println(err)
		} else {
			fmt.Fprintf(w, "Image uploaded successfully: %v \n", fileName)
		}
	}
}

func UploadFileToS3 (s *session.Session, fileName string) (string, error) {

	tempFileName := "images/" + fileName
	s3url := "https://" + os.Getenv("S3_BUCKET") + "." + os.Getenv("S3_REGION") + ".amazonaws.com/" + tempFileName

	file, erro := os.Open("./" + tempFileName)
	if erro != nil {
		return "", erro
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(os.Getenv("S3_BUCKET")),
		Key:                  aws.String("/" + tempFileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})

	if err != nil {
		return "", err
	}

	image := &models.Image{
		Name:fileName,
		S3Url:s3url,
	}
	resp := image.Create()

	SendMessage(tempFileName, s, resp)

	fmt.Println(resp)

	return tempFileName, err
}

func randomFilename (extension string) string {
	b := make([]byte, 8)
	rand.Read(b)
	fileName := hex.EncodeToString(b) + extension
	return fileName
}