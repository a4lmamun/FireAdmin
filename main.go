package main

import (
	"context"
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

var tpl *template.Template
var ctx context.Context
var app *firebase.App

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/submit", insert)
	r.Handle("/favicon.ico", http.NotFoundHandler())
	log.Fatal(http.ListenAndServe(":8080", r))
}

func index(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.gohtml", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func insert(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		password := r.FormValue("password")

		image, handle, err := r.FormFile("image")
		if err != nil {
			log.Fatalf("error image: %v\n", err)
		}
		defer image.Close()

		config := &firebase.Config{
			StorageBucket: "noteebn.appspot.com",
		}
		opt := option.WithCredentialsFile("admin.json")
		ctx = context.Background()

		app, err = firebase.NewApp(ctx, config, opt)
		if err != nil {
			log.Fatalf("error initializing app: %v\n", err)
		}

		uploadFile(handle)

		client, err := app.Auth(ctx)
		if err != nil {
			log.Fatalf("error getting Auth client: %v\n", err)
		}
		params := (&auth.UserToCreate{}).
			Email(email).
			EmailVerified(false).
			PhoneNumber(phone).
			Password(password).
			DisplayName(name).
			PhotoURL("https://firebasestorage.googleapis.com/v0/b/noteebn.appspot.com/o/" + handle.Filename + "?alt=media&token=8c18cf61-44bb-408c-bf98-b6007cee8597").
			Disabled(false)
		u, err := client.CreateUser(ctx, params)
		if err != nil {
			log.Fatalf("error creating user: %v\n", err)
		}
		log.Printf("Successfully created user: %v\n", u)
	}
}

func uploadFile(handle *multipart.FileHeader) {
	client, err := app.Storage(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		log.Fatalln(err)
	}

	file, err := handle.Open()
	if err != nil {
		log.Fatalln(err)
	}

	wc := bucket.Object(handle.Filename).NewWriter(ctx)
	if _, err = io.Copy(wc, file); err != nil {
		log.Fatalln(err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalln(err)
	}
}
