package filename

import (
	"fmt"
	"testing"
)

// Tests LinuxEncode
func TestLinuxEncode(t *testing.T) {
	inputAndExptected := map[string]string{
		"asdf/":               "asdf%2f",
		"asdf" + string(0x00): "asdf%00",
		"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfa/": "",
	}

	for inp, expected := range inputAndExptected {
		output, err := LinuxEncode(inp)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength) {
				t.Errorf("output `%v` != expected `%v`", err.Error(), fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength))
			}
		}
		if output != expected {
			t.Errorf("output `%v` != expected `%v`", output, expected)
		}
	}
}

// Tests LinuxStrip
func TestLinuxStrip(t *testing.T) {
	inputAndExptected := map[string]string{
		"asdf/":               "asdf",
		"asdf" + string(0x00): "asdf",
		"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdf": "",
	}

	for inp, expected := range inputAndExptected {
		output, err := LinuxStrip(inp)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength) {
				t.Errorf("output `%v` != expected `%v`", err.Error(), fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength))
			}
			continue
		}
		if output != expected {
			t.Errorf("output `%v` != expected `%v`", output, expected)
		}
	}
}

// Tests LinuxReplace
func TestLinuxReplace(t *testing.T) {
	var testTable = []struct {
		in struct {
			filename    string
			replacement string
		}
		out string
	}{
		{{"asdf/", "?"}, "asdf?"},
		{{"asdf" + string(0x00), "?"}, "asdf?"},
		{{"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfas/", "qwerty"}, ""},
	}

	for inp, expected := range testTable {
		output, err := LinuxReplace(inp.filename, inp.replacement)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 260, Ext4MaxLength) {
				t.Errorf("output `%v` != expected `%v`", err.Error(), fmt.Sprintf(ErrInvalidFileNameLength, 260, Ext4MaxLength))
			}
		}
		if output != expected {
			t.Errorf("output `%v` != expected `%v`", output, expected)
		}
	}
}
