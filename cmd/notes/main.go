package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

func main() {
	os.Setenv("_CLI_ZSH_AUTOCOMPLETE_HACK", "1")

	app, err := New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = app.Run(os.Args)
	if err != nil {
		if app.logger != nil {
			app.logger.Error("Failed to run command", zap.Error(err), zap.Strings("args", os.Args))
		}

		fmt.Fprintln(app.ErrWriter, err)
		os.Exit(1)
	}
}
