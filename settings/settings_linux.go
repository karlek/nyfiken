package settings

import (
	"os"
)

func setNyfikenRoot() {
	NyfikenRoot = os.Getenv("HOME") + "/.config/nyfiken/"
}
