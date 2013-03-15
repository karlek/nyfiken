// Package filename filters disallowed characters from strings to make them
// usable as filenames.
package filename

// Removes linux filename hostile characters.
// Blacklisted:
//	`/`
//	`\0`
func Linux(unencoded string) (clean string) {
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
	return clean
}
