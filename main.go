package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var scopes = []string{
	"https://www.googleapis.com/auth/youtube",
}

type CredsStruct struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func main() {
	credz := flag.String("credsFile", "", "credentials file in format of {\"client_id\": \"something\", \"client_secret\"}")
	flag.Parse()

	f, err := os.Open(*credz)
	if err != nil {
		fmt.Println(err.Error())
		panic("unable to open gcloud credentials file" + err.Error())
	}

	var creds CredsStruct
	err = json.NewDecoder(f).Decode(&creds)
	if err != nil {
		fmt.Println(err.Error())
		panic("unable to decode credentials file")
	}

	conf := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		RedirectURL:  "http://localhost:8080/thing",
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	authCodeChan := make(chan string)
	go waitforthething(authCodeChan)

	url := conf.AuthCodeURL("stsidofjosjfkjsailfkljsadfate")
	fmt.Printf("visit the url for the auth dialog %v\n", url)

	var code string
	select {
	case code = <-authCodeChan:
	case <-time.After(20 * time.Second):
		fmt.Println("too slow")
	}

	ctx := context.Background()
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		panic("unable to exchange")
	}

	client := conf.Client(ctx, tok)

	ytService, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		panic("unable to get yt service: " + err.Error())
	}

	c := ytService.Videos.List([]string{"id"})
	c.Chart("chartUnspecified")

	vlr, err := c.Do()
	if err != nil {
		panic(err.Error())
	}

	json.NewEncoder(os.Stdout).Encode(vlr)

}

func waitforthething(c chan string) {
	http.HandleFunc("/thing", func(_ http.ResponseWriter, r *http.Request) {
		c <- r.URL.Query().Get("code")
	})
	http.ListenAndServe(":8080", nil)
}
