package program

import (
	"omega/internal/env"
	"omega/internal/languages"
	"omega/internal/log"
	"omega/internal/structs"
)

func SetupLanguage(l structs.Language) {
	switch l {
	case structs.LanguageGo:
		languages.SetupGo()
	case structs.LanguageNone:

		err := structs.Config{
			LogLevel: "info",
			Commands: []string{"echo \"hello world!\""},
			Ignore:   []string{".git"},
			Timeout:  100,
			Delay:    500,
		}.WriteToFile(env.ConfigPath)
		if err != nil {
			log.Fatal("Error writing config file", "error", err)
		}
	}
}
