package main

import (
	"fmt"
	"os"

	"github.com/chickenzord/go-huawei-client/pkg/hn8010ts"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

var (
	root = &cli.App{
		Description: "CLI to interact with Huawei eg8145v5 ONT",
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
		Commands: []*cli.Command{
			{
				Name:        "opticinfo",
				Description: "Show optic info",
				Action:      opticInfo,
			},
			{
				Name:        "top",
				Description: "Show router resources usage",
				Action:      top,
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

func opticInfo(ctx *cli.Context) error {
	cfg := &hn8010ts.Config{
		URL:      ctx.String("url"),
		Username: ctx.String("username"),
		Password: ctx.String("password"),
	}

	client := hn8010ts.NewClient(*cfg)

	if err := client.Session(func(c *hn8010ts.Client) error {
		opticInfo, err := c.GetOpticInfo()
		if err != nil {
			return err
		}

		fmt.Printf("RX Level: %.2f TX Level: %.2f\n", opticInfo.RXPower, opticInfo.TXPower)

		return nil
	}); err != nil {
		return err
	}

	return nil
}
func top(ctx *cli.Context) error {
	cfg := &hn8010ts.Config{
		URL:      ctx.String("url"),
		Username: ctx.String("username"),
		Password: ctx.String("password"),
	}

	client := hn8010ts.NewClient(*cfg)

	if err := client.Session(func(c *hn8010ts.Client) error {
		usage, err := c.GetResourceUsage()
		if err != nil {
			return err
		}

		fmt.Printf("Memory: %d%%\n", usage.Memory)
		fmt.Printf("CPU: %d%%\n", usage.CPU)

		return nil
	}); err != nil {
		return err
	}

	return nil
}
