// +build !debug

package main

import (
	"github.com/subtlepseudonym/notes"
	dalpkg "github.com/subtlepseudonym/notes/dal"

	"github.com/urfave/cli"
)

func buildDebugCommand(dal dalpkg.DAL, meta *notes.Meta) cli.Command {
	return cli.Command{
		Hidden: true,
	}
}
