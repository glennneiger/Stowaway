package main

import (
	"Stowaway/admin"
	"Stowaway/agent"
	"fmt"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const version = "0.0.1"

var (
	timestamp = ""
	githash   = ""
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go listenInterrupt(c)
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("hello")
	}
	app := &cli.App{}
	app.Name = "stowaway"
	app.Commands = []*cli.Command{
		agent.Command,
		admin.Command,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func listenInterrupt(c chan os.Signal) {
	<-c
	os.Exit(1)
}
