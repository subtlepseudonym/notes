# Changelog

## [1.1.0]
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
