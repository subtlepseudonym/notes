package dal

// IsHidden determines if a file is hidden for a unix system
//
// File names may not be empty as of the unix specification, so
// we're likely safe assuming that len(filename) > 0
func IsHidden(filename string) bool {
	return filename[0] == '.'
}
