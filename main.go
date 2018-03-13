package main

import (
	"errors"
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	version = "0.0.1"
)

func main() {
	var err error
	app := NewApp()
	if err != nil {
		app.log.Fatal(err)
	}

	a := cli.NewApp()
	a.Version = version
	a.Usage = ""
	a.Description = "A tool for insert cronjobs (uniquely) into a crontab"
	a.Usage = "sync journal reports to local database"
	a.Action = func(c *cli.Context) error {
		alias := c.String("alias")
		cronjobs := c.StringSlice("cronjob")

		if alias == "" || len(cronjobs) == 0 {
			err := errors.New("Both --alias and --cronjob are required")
			return cli.NewExitError(err, 1)
		}

		err := app.Run(alias, cronjobs)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	}
	a.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "alias",
			Usage: "The alias to uniquely identify the cronjob",
		},
		cli.StringSliceFlag{
			Name:  "cronjob",
			Usage: "The cronjob line itself",
		},
	}

	err = a.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
