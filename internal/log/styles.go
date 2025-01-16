package log

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var defaultStyle = func() *log.Styles {
	styles := log.DefaultStyles()

	style := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1, 0, 1)

	styles.Levels[log.DebugLevel] = style.
		SetString(TAG + "D").
		Background(lipgloss.Color("#414868")).
		Foreground(lipgloss.Color("#FFFFFF"))

	styles.Levels[log.InfoLevel] = style.
		SetString(TAG + "I").
		Background(lipgloss.Color("#485E30")).
		Foreground(lipgloss.Color("#FFFFFF"))

	styles.Levels[log.WarnLevel] = style.
		SetString(TAG + "W").
		Background(lipgloss.Color("#FF9E64")).
		Foreground(lipgloss.Color("#000000"))

	styles.Levels[log.ErrorLevel] = style.
		SetString(TAG + "E").
		Background(lipgloss.Color("#F7768E")).
		Foreground(lipgloss.Color("#000000"))

	styles.Levels[log.FatalLevel] = style.
		SetString(TAG + "F").
		Background(lipgloss.Color("#BB9AF7")).
		Foreground(lipgloss.Color("#000000"))

	return styles
}()
