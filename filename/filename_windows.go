// Package filename filters disallowed characters from strings to make them
// usable as filenames.
//
// *Warning* - Filename is not safe to use since all restrictions have not been
// added.
package filename

import (
	"fmt"
)

const (
	NTFSMaxLength = 255
)

// Characters whose integer representations are in the range from 1 through 31.

// Do not use the following reserved names for the name of a file:
// Also avoid these names followed immediately by an extension; for example, NUL.txt is not recommended.
var reservedNames = map[string]bool{
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

// Encodes windows filename hostile characters.
// Blacklisted:
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
func Encode(unencoded string) (clean string, err error) {
	for _, chr := range unencoded {
		switch chr {
		case '\\':
			clean += `%5c`
		case '/':
			clean += `%2f`
		case ':':
			clean += `%3a`
		case '*':
			clean += `%2a`
		case '?':
			clean += `%3f`
		case '"':
			clean += `%22`
		case '<':
			clean += `%3c`
		case '>':
			clean += `%3e`
		case '|':
			clean += `%7c`
		case 0x00:
			clean += `%00`
		default:
			clean += string(chr)
		}
	}

	lenClean := len(clean)
	if lenClean > 255 {
		return "", fmt.Errorf(ErrInvalidFileNameLength, lenClean, NTFSMaxLength)
	}
	return clean, nil
}

// Removes windows filename hostile characters.
// Blacklisted:
//  `\\`
//  `/`
//  `:`
//  `*`
//  `?`
//  `"`
//  `<`
//  `>`
//  `|`
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
		return "", fmt.Errorf(ErrInvalidFileNameLength, lenClean, NTFSMaxLength)
	}
	return clean, nil
}

// Replaces windows filename hostile characters with user supplied string.
// Blacklisted:
//  `\\`
//  `/`
//  `:`
//  `*`
//  `?`
//  `"`
//  `<`
//  `>`
//  `|`
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
		return "", fmt.Errorf(ErrInvalidFileNameLength, lenClean, NTFSMaxLength)
	}
	return clean, nil
}

func IsRestricted(r rune) bool {
	switch r {
	case '\\', '/', ':', '*', '?', '"', '<', '>', '|', string(0x00):
		return true
	default:
		return false
	}
}

func IsLastSpaceOrPeriod(filename string) bool {
	if r := filename[:len(filename)-1]; r == ' ' || r == '.' {
		return true
	}
	return false
}

func IsRestrictedFilename(filename string) bool {
	found, ok := reservedNames[filename]
	return ok
}
