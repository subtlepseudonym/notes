package dal

import "syscall"

// IsHidden determines if a file is hidden for a windows system
// It swallows errors in accessing the file attributes and returns
// false (file is not hidden) if there is an error
func IsHidden(filename string) bool {
	ptr, err := syscall.UTF16PointerFromString(filename)
	if err != nil {
		return false
	}

	attr, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		return false
	}

	hidden := attr & syscall.FILE_ATTRIBUTE_HIDDEN
	return hidden, nil
}
