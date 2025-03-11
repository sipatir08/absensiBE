package firebase

import (
    "context"
    "log"
    "os"

    firebase "firebase.google.com/go"
    "firebase.google.com/go/auth"
    "google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitFirebase() *auth.Client {
    credsPath := os.Getenv("FIREBASE_CREDENTIALS")
    if credsPath == "" {
        log.Fatal("FIREBASE_CREDENTIALS not set in .env")
    }

    opt := option.WithCredentialsFile(credsPath)
    app, err := firebase.NewApp(context.Background(), nil, opt)
    if err != nil {
        log.Fatal("Error initializing Firebase App:", err)
    }

    client, err := app.Auth(context.Background())
    if err != nil {
        log.Fatal("Error initializing Firebase Auth:", err)
    }

    AuthClient = client
    return client
}