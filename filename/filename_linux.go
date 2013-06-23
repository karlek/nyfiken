// Linux blacklisted characters:
//
//		`/`
//		`\0`
package filename

import (
	"fmt"
	"net/url"
)

const (
	Ext4MaxLength = 255 // Longest allowed filename on ext-4.
)

// Encode encodes linux filename hostile characters from string unsafe with
// percent-encoding.
func Encode(unsafe string) (clean string, err error) {
	for _, chr := range unsafe {
		switch {
		case IsHostile(chr):
			clean += url.QueryEscape(string(chr))
		default:
			clean += string(chr)
		}
	}
	if !IsSafeLen(clean) {
		return "", fmt.Errorf(ErrInvalidFileNameLength, len(clean), Ext4MaxLength)
	}
	return clean, nil
}

// IsSafeLen verifies if string s has an allowed filename length.
func IsSafeLen(s string) bool {
	return len(s) < Ext4MaxLength
}

// Strip removes linux filename hostile characters from string unsafe.
func Strip(unsafe string) (clean string, err error) {
	for _, chr := range unsafe {
		switch {
		case IsHostile(chr):
			continue
		default:
			clean += string(chr)
		}
	}
	if !IsSafeLen(clean) {
		return "", fmt.Errorf(ErrInvalidFileNameLength, len(clean), Ext4MaxLength)
	}
	return clean, nil
}

// Replace replaces linux filename hostile characters in string unsafe with string replace.
func Replace(unsafe, replace string) (clean string, err error) {
	for _, chr := range unsafe {
		switch {
		case IsHostile(chr):
			clean += replace
		default:
			clean += string(chr)
		}
	}
	if !IsSafeLen(clean) {
		return "", fmt.Errorf(ErrInvalidFileNameLength, len(clean), Ext4MaxLength)
	}
	return clean, nil
}

// IsHostile verifies if rune r is a linux filename hostile characters.
func IsHostile(r rune) bool {
	return r == '/' || r == 0x00
}
