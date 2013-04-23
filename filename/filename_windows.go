// Package filename filters disallowed characters from strings to make them
// usable as filenames.
//
// *Warning* - Filename is not safe to use since all restrictions have not been
// added.
package filename

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mewkiz/pkg/errutil"
)

const (
	NTFSMaxLength = 255
)

var (
	ErrLastCharIsSpaceOrPeriod = errors.New("last char is space or period.")
	ErrNTFSRestrictedFilename  = errors.New("filename is restricted in NTFS.")

	// Do not use the following reserved names for the name of a file:
	// Also avoid these names followed immediately by an extension;
	// for example, NUL.txt is not recommended.
	reservedNames = map[string]bool{
		"CON":  true,
		"PRN":  true,
		"AUX":  true,
		"NUL":  true,
		"COM1": true,
		"COM2": true,
		"COM3": true,
		"COM4": true,
		"COM5": true,
		"COM6": true,
		"COM7": true,
		"COM8": true,
		"COM9": true,
		"LPT1": true,
		"LPT2": true,
		"LPT3": true,
		"LPT4": true,
		"LPT5": true,
		"LPT6": true,
		"LPT7": true,
		"LPT8": true,
		"LPT9": true,
	}
)

// Encodes NTFS filename hostile characters.
func Encode(unencoded string) (clean string, err error) {
	for _, chr := range unencoded {
		switch chr {
		case IsRestricted(r):
			clean += `%` + fmt.Sprintf("%x", r)
		default:
			clean += string(r)
		}
	}

	err = IsAllowed(clean)
	if err != nil {
		return errutil.Err(err)
	}

	return clean, nil
}

// Removes NTFS filename hostile characters.
func Strip(unencoded string) (clean string, err error) {
	for _, chr := range unencoded {
		switch chr {
		case IsRestricted(r):
			continue
		default:
			clean += string(chr)
		}
	}

	lenClean := len(clean)
	if lenClean > 255 {
		return "", errutil.Newf(ErrInvalidFileNameLength, lenClean, NTFSMaxLength)
	}
	return clean, nil
}

// Replaces NTFS filename hostile characters with user supplied string.
func Replace(unencoded, replace string) (clean string, err error) {
	for _, r := range unencoded {
		switch r {
		case IsRestricted(r):
			clean += replace
		default:
			clean += string(r)
		}
	}

	lenClean := len(clean)
	if lenClean > 255 {
		return "", errutil.Newf(ErrInvalidFileNameLength, lenClean, NTFSMaxLength)
	}
	return clean, nil
}

// NTFS restricted characters:
//  `\\`
//  `/`
//  `:`
//  `*`
//  `?`
//  `"`
//  `<`
//  `>`
//  `|`
//  `\0`
// 		and
// characters whose integer representations are in the range from 1 through 31.
func IsRestricted(r rune) bool {
	switch r {
	case '\\', '/', ':', '*', '?', '"', '<', '>', '|', string(0x00):
		return true
	case r >= string(0x00) && r <= string(0x1f):
		return true
	default:
		return false
	}
}

// Filenames are not allowed to end in ` ` (%20 i.e space) or `.`
// (%2e i.e period).
func IsLastSpaceOrPeriod(filename string) bool {
	if r := filename[:len(filename)-1]; r == ' ' || r == '.' {
		return true
	}
	return false
}

// Filenames aren't allowed to be certain reserved filenames.
func IsRestrictedFilename(filename string) bool {
	_, ok := reservedNames[filepath.Base(filename)]
	return ok
}

// IsAllowed returns if a string is usable as a filename.
func IsAllowed(filename string) (err error) {
	lenClean := len(filename)
	switch {
	case lenClean > 255:
		return errutil.Newf(ErrInvalidFileNameLength, lenClean, NTFSMaxLength)
	case IsLastSpaceOrPeriod(filename):
		return errutil.Err(ErrLastCharIsSpaceOrPeriod)
	case IsRestrictedFilename(filename):
		return errutil.Err(ErrNTFSRestrictedFilename)
	}
	return nil
}
