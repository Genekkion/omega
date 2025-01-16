package languages

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"omega/internal/env"
	"omega/internal/log"
	"omega/internal/structs"
	"os"
	"os/exec"
	"path/filepath"
	"unicode"

	"golang.org/x/mod/modfile"

	"github.com/charmbracelet/huh"
)

//go:embed main.go.template
var mainFileGo string

func SetupGo() {
	moduleName := checkGoMod()

	config := structs.Config{
		Ignore:   []string{".git", "bin"},
		LogLevel: "info",
		Timeout:  100,
		Commands: []string{},
		Delay:    500,
	}

	binPath := filepath.Join("bin", moduleName)
	config.Commands = []string{
		"go build -o " + binPath + " ./cmd",
		binPath,
	}

	err := createGoCmd()
	if err != nil {
		log.Warn("Error occurred creating cmd folder, skipping", "error", err)
		config.Commands = []string{}
	} else {
		err = createGoMain(moduleName)
		if err != nil {
			log.Warn("Error occurred creating cmd/main.go, skipping", "error", err)
			config.Commands = []string{}
		}
	}

	config.WriteToFile(env.ConfigPath)
	if err != nil {
		log.Fatal("Error creating config file", "error", err)
	}
}

// Create a cmd folder
func createGoCmd() error {
	err := os.Mkdir("cmd", 0755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		log.Fatal("Error occurred making directory", "error", err)
	}
	return err
}

// Create cmd/main.go
func createGoMain(moduleName string) error {
	file, err := os.Create(filepath.Join("cmd", "main.go"))
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(fmt.Sprintf(mainFileGo, moduleName)))
	return err
}

// Extracts the module name from an existing go.mod file
func parseGoMod(filePath string, file *os.File) (string, error) {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	modFile, err := modfile.Parse(filePath, bytes, nil)
	if err != nil {
		return "", err
	}
	return modFile.Module.Mod.String(), nil
}

func goModuleValidator(str string) error {
	if str == "" {
		return errors.New("Cannot be left empty!")
	}
	for i, r := range str {
		if i == 0 && !unicode.In(r, unicode.Letter) {
			return errors.New("Needs to start with alphabet!")
		} else if !unicode.In(r, unicode.Letter, unicode.Number) {
			switch r {
			case '-', '_', '/', '.':
				continue
			default:
				return fmt.Errorf("Invalid rune! \"%c\"", r)
			}
		}
	}
	return nil
}

func formGoModuleName() (string, error) {
	var moduleName string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Couldn't find a go.mod file, pick a name for this project").
			Validate(goModuleValidator).
			Value(&moduleName),
	)).Run()
	if err != nil {
		return "", err
	}
	return moduleName, nil
}

// Returns the module name found, or created by user
func checkGoMod() string {
	// Attempts to look for the go.mod file in the same directory
	filePath := filepath.Join(".", "go.mod")
	file, err := os.Open(filePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal("Error opening the existing module file at "+filePath, "error", err)
	} else if err != nil {
		// File does not exist
		file, err = os.Create(filePath)
		if err != nil {
			log.Fatal("Error creating a module file at "+filePath, "error", err)
		}
	}

	defer file.Close()
	if err == nil {
		moduleName, err := parseGoMod(filePath, file)
		if err != nil {
			log.Fatal("Error parsing module file at "+filePath, "error", err)
		}
		return moduleName
	}

	moduleName, err := formGoModuleName()
	if err != nil {
		log.Fatal("Error running form :(", "error", err)
	}

	cmd := exec.Command("go", "mod", "init", moduleName)
	err = cmd.Run()
	if err != nil {
		log.Fatal("Error occurred creating go.mod file", "error", err)
	}
	return moduleName
}
