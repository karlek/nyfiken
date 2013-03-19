// Package filename filters disallowed characters from strings to make them
// usable as filenames.
package filename

import (
	"fmt"
)

const (
	Ext4MaxLength = 255
)

// Encodes linux filename hostile characters.
// Blacklisted:
//	`/`
//	`\0`
func Encode(unencoded string) (clean string, err error) {
	for _, chr := range unencoded {
		switch chr {
		case '/':
			clean += `%2f`
		case 0x00:
			clean += "%00"
		default:
			clean += string(chr)
		}
	}

	lenClean := len(clean)
	if lenClean > 255 {
		return "", fmt.Errorf(ErrInvalidFileNameLength, lenClean, Ext4MaxLength)
	}
	return clean, nil
}

// Removes linux filename hostile characters.
// Blacklisted:
//	`/`
//	`\0`
func Strip(unencoded string) (clean string, err error) {
	for _, chr := range unencoded {
		switch chr {
		case '/', 0x00:
			continue
		default:
			clean += string(chr)
		}
	}

	lenClean := len(clean)
	if lenClean > 255 {
		return "", fmt.Errorf(ErrInvalidFileNameLength, lenClean, Ext4MaxLength)
	}
	return clean, nil
}

// Replaces linux filename hostile characters with user supplied string.
// Blacklisted:
//	`/`
//	`\0`
func Replace(unencoded, replace string) (clean string, err error) {
	for _, chr := range unencoded {
		switch chr {
		case '/', 0x00:
			clean += replace
		default:
			clean += string(chr)
		}
	}

	lenClean := len(clean)
	if lenClean > 255 {
		return "", fmt.Errorf(ErrInvalidFileNameLength, lenClean, Ext4MaxLength)
	}
	return clean, nil
}
