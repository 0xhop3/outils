package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseAuthentication struct {
	client *auth.Client
}

func NewFirebaseAuthentication(credentialFile string) (*FirebaseAuthentication, error) {
	opt := option.WithCredentialsFile(credentialFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication client: %w", err)
	}

	return &FirebaseAuthentication{client: client}, nil
}

func (f *FirebaseAuthentication) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil
}
