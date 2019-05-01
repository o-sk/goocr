package goocr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	config  *Config
	Client  *http.Client
	Service *drive.Service
}

func NewGoocr(config *Config) *Goocr {
	return &Goocr{
		config: config,
	}
}

func (g *Goocr) SetupClient() (err error) {
	b, err := ioutil.ReadFile(g.config.credentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "Unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config.")
	}
	g.Client, err = g.getClient(config)
	if err != nil {
		return errors.Wrap(err, "Failed new drive service.")
	}

	g.Service, err = drive.New(g.Client)
	if err != nil {
		return errors.Wrap(err, "Failed new drive service.")
	}

	return nil
}

func (g *Goocr) Recognize(path string) (text string, err error) {
	uploadFile, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "Can't open %s.", path)
	}

	driveFile, err := g.upload(uploadFile)
	if err != nil {
		return "", err
	}

	text, err = g.read(driveFile)
	if err != nil {
		return "", err
	}

	err = g.delete(driveFile)
	if err != nil {
		return "", err
	}

	return text, err
}

func (g *Goocr) upload(file *os.File) (driveFile *drive.File, err error) {
	f := &drive.File{Name: file.Name(), MimeType: "application/vnd.google-apps.document"}
	driveFile, err = g.Service.Files.Create(f).Media(file).Do()
	if err != nil {
		return nil, errors.Wrap(err, "Failed upload")
	}
	return driveFile, nil
}

func (g *Goocr) read(driveFile *drive.File) (string, error) {
	res, err := g.Service.Files.Export(
		driveFile.Id,
		"text/plain",
	).Download()
	if err != nil {
		return "", err
	}
	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (g *Goocr) delete(driveFile *drive.File) (err error) {
	err = g.Service.Files.Delete(driveFile.Id).Do()
	if err != nil {
		return errors.Wrap(err, "Failed delete")
	}
	return nil
}

func (g *Goocr) getClient(config *oauth2.Config) (*http.Client, error) {
	tok, err := g.tokenFromFile(g.config.tokenFilePath)
	if err != nil {
		tok, err = g.getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		err = g.saveToken(g.config.tokenFilePath, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
}

func (g *Goocr) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.Wrap(err, "Unable to read authorization code.")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to retrieve token from web.")
	}
	return tok, nil
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

func (g *Goocr) saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "Unable to cache oauth token.")
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}
