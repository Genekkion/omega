package forms

import (
	"omega/internal/log"

	"github.com/charmbracelet/huh"
)

func GenerateConfig() bool {
	var flag bool
	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Would you like to create a new config file?").
			Value(&flag),
	)).Run()
	if err != nil {
		log.Fatal("An error running the form :(", "error", err)
	}
	return flag
}
