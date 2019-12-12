// +build !debug

package main

import (
	"github.com/urfave/cli"
)

func (a *App) buildDebugCommand() cli.Command {
	return cli.Command{
		Hidden: true,
	}
}
