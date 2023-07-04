package main

import (
	"context"
	"fmt"
	"os"

	"github.com/livebud/cli"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cli := cli.New("duo", "duo templating language")

	{ // serve [flags] [dir]
		cmd := new(Serve)
		cli := cli.Command("serve", "serve a directory")
		cli.Flag("listen", "address to listen on").String(&cmd.Listen).Default(":3000")
		cli.Flag("live", "enable live reloading").Bool(&cmd.Live).Default(true)
		cli.Flag("open", "open browser").Bool(&cmd.Browser).Default(true)
		cli.Arg("dir").String(&cmd.Dir).Default(".")
		cli.Run(cmd.Run)
	}

	return cli.Parse(context.Background(), os.Args[1:]...)
}

type Serve struct {
	Listen  string
	Live    bool
	Dir     string
	Browser bool
}

func (s *Serve) Run(ctx context.Context) error {
	return fmt.Errorf("not implemented yet")
}
