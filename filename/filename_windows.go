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
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
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
	for _, chr := range unencoded {
		switch chr {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			clean += replace
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
