package embed

import (
	"embed"
	"sync"
)

var (
	//go:embed usagetheme.json USAGE.md
	files embed.FS

	once sync.Once
	ef   *EmbeddedFiles
)

type EmbeddedFiles struct {
	UsageTheme []byte
	UsageFile  []byte
}

func EmbeddedFilesInstance() *EmbeddedFiles {
	once.Do(func() {
		usageTheme, err := files.ReadFile("usagetheme.json")
		if err != nil {
			panic(err) // it's for developer to ensure no error occurs during prod
		}
		usageFile, err := files.ReadFile("USAGE.md")
		if err != nil {
			panic(err)
		}
		ef = &EmbeddedFiles{usageTheme, usageFile}
	})
	return ef
}
