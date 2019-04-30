package goocr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
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

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	g.Client = g.getClient(config)
}

func (g *Goocr) Recognize(path string) (text string, err error) {
	uploadFile, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "Can't open %s.", path)
	}

	service, err := drive.New(g.Client)
	if err != nil {
		return "", errors.Wrap(err, "Failed new drive service.")
	}

	driveFile, err := g.upload(service, uploadFile)
	if err != nil {
		return "", err
	}

	err = g.delete(service, driveFile)
	if err != nil {
		return "", err
	}

	return text, err
}

func (g *Goocr) upload(service *drive.Service, file *os.File) (driveFile *drive.File, err error) {
	f := &drive.File{Name: file.Name()}
	driveFile, err = service.Files.Create(f).Media(file).Do()
	if err != nil {
		return nil, errors.Wrap(err, "Failed upload")
	}
	return driveFile, nil
}

func (g *Goocr) delete(service *drive.Service, driveFile *drive.File) (err error) {
	err = service.Files.Delete(driveFile.Id).Do()
	if err != nil {
		return errors.Wrap(err, "Failed delete")
	}
	return nil
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
