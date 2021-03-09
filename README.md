## Notes

This project is intended to make it easy to jot down notes and retrieve them in an organized fashion at a later date. It's targeted at people who spend a lot of time in the command line.

### Development

When running the recipes included in the makefile, you may want to have `vtag` in your PATH. The recipes will work without it, but the resulting binary will not include version information. You can get `vtag` by cloning [subtlepseudonym/utilities](https://github.com/subtlepseudonym/utilities), running `make build`, and copying `bin/vtag` into your PATH.

You can build the notes binary with `make build`. This will download dependencies and assign some meta information, including version, to the binary. You can then access this info with `notes info` after the build is complete.
