package main

import (
	"fmt"
	"os"

	"github.com/chickenzord/go-huawei-client/pkg/eg8145v5"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var (
	root = &cli.App{
		Commands: []*cli.Command{
			devices,
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "url",
				EnvVars: []string{"ROUTER_URL"},
			},
			&cli.StringFlag{
				Name:    "username",
				Aliases: []string{"u"},
				EnvVars: []string{"ROUTER_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				EnvVars: []string{"ROUTER_PASSWORD"},
			},
		},
	}
	devices = &cli.Command{
		Name: "devices",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Action: devicesList,
			},
		},
	}
)

func main() {
	_ = godotenv.Overload()

	if err := root.Run(os.Args); err != nil {
		fmt.Println()
		fmt.Println(err)
		os.Exit(1)
	}
}

func devicesList(ctx *cli.Context) error {
	cfg := &eg8145v5.Config{
		URL:      ctx.String("url"),
		Username: ctx.String("username"),
		Password: ctx.String("password"),
	}

	client := eg8145v5.NewClient(*cfg)

	if err := client.Session(func(c *eg8145v5.Client) error {
		devices, err := c.ListUserDevices()
		if err != nil {
			return err
		}

		for _, d := range devices {
			fmt.Println(d.HostName, d.DevStatus)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
