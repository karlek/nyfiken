package settings

import (
	"os"
)

func setNyfikenRoot() {
	NyfikenRoot = os.Getenv("APPDATA") + "/nyfiken"
}
