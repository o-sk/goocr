package main

import (
	"fmt"
	"log"
	"os"

	"github.com/o-sk/goocr/goocr"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "goocr"
	app.Usage = "Optical character recognition with google drive"
	app.Version = "0.0.1"

	var (
		credentialsFilePath string
		tokenFilePath       string
		objectFilePath      string
	)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "credentials, c",
			Usage:       "Credentials file path",
			Value:       "credentials.json",
			Destination: &credentialsFilePath,
		},
		cli.StringFlag{
			Name:        "token, t",
			Usage:       "Token file path",
			Value:       "token.json",
			Destination: &tokenFilePath,
		},
		cli.StringFlag{
			Name:        "fle, f",
			Usage:       "Object file path",
			Value:       "",
			Destination: &objectFilePath,
		},
	}

	app.Action = func(context *cli.Context) error {
		if objectFilePath == "" {
			log.Fatal("Not given file path")
			return nil
		}
		g := goocr.NewGoocr(goocr.NewConfig(credentialsFilePath, tokenFilePath))
		err := g.SetupClient()
		if err != nil {
			log.Fatalf("Can't setup client. %v", err)
			return nil
		}
		text, err := g.Recognize(objectFilePath)
		if err != nil {
			log.Fatalf("Can't recognize. %v", err)
			return nil
		}
		fmt.Printf("%s\n", text)
		return nil
	}
	app.Run(os.Args)
}
