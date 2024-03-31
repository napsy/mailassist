package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type mailProvider interface {
	fetch() ([]providerMessage, error)
}

type providerMessage struct {
	header  map[string]string
	message string
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type gmailProvider struct {
	prefetchN int
	service   *gmail.Service
}

func newGmailProvider(credsFile string, prefetchN int) (*gmailProvider, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}
	return &gmailProvider{
		prefetchN: prefetchN,
		service:   srv,
	}, nil
}

func (gmail *gmailProvider) fetch() ([]providerMessage, error) {
	user := "me"
	r, err := gmail.service.Users.Messages.List(user).Q("is:unread").MaxResults(int64(gmail.prefetchN)).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	msgs := []providerMessage{}
	for _, m := range r.Messages {
		msg, err := gmail.service.Users.Messages.Get(user, m.Id).Format("full").Do()
		if err != nil {
			continue
		}
		// Parse the internalDate of the message to check if it's within the last 30 minutes
		internalDate := time.Unix(0, msg.InternalDate*int64(time.Millisecond))
		if time.Since(internalDate).Minutes() > 30 {
			continue
		}
		if msg.Payload.Body.Size > 0 {
			header := make(map[string]string)
			for _, h := range msg.Payload.Headers {
				header[h.Name] = h.Value
			}
			msgs = append(msgs, providerMessage{header: header, message: msg.Payload.Body.Data})
		} else {
			for _, part := range msg.Payload.Parts {
				header := make(map[string]string)
				for _, h := range msg.Payload.Headers {
					header[h.Name] = h.Value
				}
				msgs = append(msgs, providerMessage{header: header, message: part.Body.Data})
			}
		}
	}
	return msgs, nil
}
