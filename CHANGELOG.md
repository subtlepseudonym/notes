# Changelog

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
- Make command help text more helpful (and the format was standardized)
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
