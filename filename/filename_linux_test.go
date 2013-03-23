package filename

import (
	"fmt"
	"testing"
)

// Tests Encode
func TestEncode(t *testing.T) {
	var testTable = []struct {
		filename string
		out      string
	}{
		{"asdf/", "asdf%2f"},
		{"asdf" + string(0x00), "asdf%00"},
		{"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfa/", ""},
	}

	for _, expected := range testTable {
		output, err := Encode(expected.filename)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength) {
				t.Errorf("output `%v` != expected `%v`", err.Error(), fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength))
			}
		}
		if output != expected.out {
			t.Errorf("output `%v` != expected `%v`", output, expected.out)
		}
	}
}

// Tests Strip
func TestStrip(t *testing.T) {
	var testTable = []struct {
		filename string
		out      string
	}{
		{"asdf/", "asdf"},
		{"asdf" + string(0x00), "asdf"},
		{"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdf", ""},
	}

	for _, expected := range testTable {
		output, err := Strip(expected.filename)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength) {
				t.Errorf("output `%v` != expected `%v`", err.Error(), fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength))
			}
			continue
		}
		if output != expected.out {
			t.Errorf("output `%v` != expected `%v`", output, expected.filename)
		}
	}
}

// Tests Replace
func TestReplace(t *testing.T) {
	var testTable = []struct {
		filename    string
		replacement string
		out         string
	}{
		{"asdf/", "?", "asdf?"},
		{"asdf" + string(0x00), "?", "asdf?"},
		{"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfas/", "qwerty", ""},
	}

	for _, expected := range testTable {
		output, err := Replace(expected.filename, expected.replacement)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 260, Ext4MaxLength) {
				t.Errorf("output `%v` != expected `%v`", err.Error(), fmt.Sprintf(ErrInvalidFileNameLength, 260, Ext4MaxLength))
			}
		}
		if output != expected.out {
			t.Errorf("output `%v` != expected `%v`", output, expected.out)
		}
	}
}
