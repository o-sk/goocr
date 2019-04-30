package goocr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type Config struct {
	credentialsFilePath string
	tokenFilePath       string
}

func NewConfig(credentialsFilePath, tokenFilePath string) *Config {
	return &Config{
		credentialsFilePath: credentialsFilePath,
		tokenFilePath:       tokenFilePath,
	}
}

type Goocr struct {
	config *Config
	Client *http.Client
}

func NewGoocr(config *Config) *Goocr {
	return &Goocr{
		config: config,
	}
}

func (g *Goocr) SetupClient() {
	b, err := ioutil.ReadFile(g.config.credentialsFilePath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveMetadataScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	g.Client = g.getClient(config)
}

func (g *Goocr) getClient(config *oauth2.Config) *http.Client {
	tok, err := g.tokenFromFile(g.config.tokenFilePath)
	if err != nil {
		tok = g.getTokenFromWeb(config)
		g.saveToken(g.config.tokenFilePath, tok)
	}
	return config.Client(context.Background(), tok)
}

func (g *Goocr) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func (g *Goocr) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (g *Goocr) saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
