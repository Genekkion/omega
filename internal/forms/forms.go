package forms

import (
	"omega/internal/log"
	"omega/internal/structs"

	"github.com/charmbracelet/huh"
)

func FormGenerateConfig() bool {
	var generateConfig bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Would you like to create a new config file?").
			Value(&generateConfig),
	)).Run()
	if err != nil {
		log.Fatal("Error running the form :(", "error", err)
	}
	return generateConfig
}

func FormSelectLanguage() structs.Language {
	options := make([]huh.Option[structs.Language], 0, len(structs.LanguageMap))
	for k, v := range structs.LanguageMapInverted {
		options = append(options, huh.NewOption(k, v))
	}

	var l structs.Language
	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[structs.Language]().
			Title("Select a language").
			Options(options...).
			Value(&l),
	)).Run()
	if err != nil {
		log.Fatal("Error running the form :(", "error", err)
	}

	return l
}
