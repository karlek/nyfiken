// Test cases for filename
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

	for inp, exptected := range inputAndExptected {
		output, err := LinuxEncode(inp)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength) {
				t.FailNow()
			}
		}
		if output != exptected {
			t.FailNow()
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

	for inp, exptected := range inputAndExptected {
		output, err := LinuxStrip(inp)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 256, Ext4MaxLength) {
				t.FailNow()
			}
			continue
		}
		if output != exptected {
			t.FailNow()
		}
	}
}

// Tests LinuxReplace
func TestLinuxReplace(t *testing.T) {
	type input struct {
		filename    string
		replacement string
	}
	inputAndExptected := map[input]string{
		input{"asdf/", "?"}:               "asdf?",
		input{"asdf" + string(0x00), "?"}: "asdf?",
		input{"asdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfasdfas/", "qwerty"}: "",
	}

	for inp, exptected := range inputAndExptected {
		output, err := LinuxReplace(inp.filename, inp.replacement)
		if err != nil {
			if err.Error() != fmt.Sprintf(ErrInvalidFileNameLength, 260, Ext4MaxLength) {
				t.FailNow()
			}
		}
		if output != exptected {
			t.FailNow()
		}
	}
}
