# Changelog

## [2.0.0] -
### Added
- Caching for note retrieval
- Debug command for getting NoteMeta objects
- Debug command for rebuilding local DAL index
- Separate 'notebooks' for partitioning notes
- Direct access to NoteMeta objects through DAL

### Changed
- Rename notes/dal package dalpkg to dal
- Index note meta information separately from overall meta in local DAL
- Stop using mitchellh/go-homedir in favor of os.UserHomeDir (introduced in go1.13.7)
- Rename dal.NewLocalDAL to dal.NewLocal

## [1.2.3] - 2020-05-02
### Fixed
- Prevent meta file falling out of sync when terminal crashes

## [1.2.2] - 2019-12-17
### Changed
- Background updater has more descriptive logging

## [1.2.1] - 2019-12-12
### Fixed
- Logging correctly uses the shared state logger

## [1.2.0] - 2019-12-12
### Added
- Generate release candidate versions for \*-rc branches in build tag script
- Loads of additional logging
- Flag for writing / editing notes without appending to the note's edit history
- Debug commands for accessing lower level functionality (debug)
	- Going forward, the debug command and its subcommands will not be considered a part of the public API
- Automatically update meta version (unless app version is a major release ahead)

### Changed
- Use zap logging
- Include main package version in logs
- Upgrade to go1.13 error wrapping
- List build tags in app info if any are present
- Generate version / build tag with a go utility rather than a script
- Deprecated notes/log package

### Fixed
- Handle errors more gracefully in interactive mode
- Initialize logger / home directory value once
- Stop using package global logger

## [1.1.2] - 2019-11-14
### Changed
- Updated to go1.13

## [1.1.1] - 2019-07-02
### Fixed
- Write meta latestId field before user handoff in case editor crashes

## [1.1.0] - 2019-06-21
### Added
- Interactive mode
- Automatic note and meta updating in the background
- Logging for background processes
- History for interactive mode

### Changed
- Make command help text more helpful (and standardize format)
- Improve `edit` command's ability to find the latest note when no ID is provided
- Provide more descriptive build tag in semantic version
- Trim exit command in interactive mode
- Avoid showing "Deleted" field in info output if note isn't deleted
- Rename DefaultDAL to LocalDAL (technically a breaking change)
- Retrieve DAL and meta when notes starts rather than when each command is run

### Fixed
- Fix revision displayed with `info` command
- Improve build tag counting of untracked files

## [1.0.0] - 2019-03-19
### Added
- List existing notes command (ls)
- Create new note command (new)
- Edit existing notes command (edit)
- Delete notes command (rm)
- Local filesystem DAL
- Note edit history
