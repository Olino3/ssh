package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	primary      = lipgloss.Color("#8B5CF6")
	successColor = lipgloss.Color("#22C55E")
	errorColor   = lipgloss.Color("#EF4444")
	warnColor    = lipgloss.Color("#F59E0B")
	mutedColor   = lipgloss.Color("#6B7280")
	box          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

func PrintBanner() {
	fmt.Println(box.BorderForeground(primary).Bold(true).Render("sshutil — ssh + tailscale, set up right"))
}

func Info(format string, a ...any) { log.Infof(format, a...) }
func Muted(format string, a ...any) {
	fmt.Println(lipgloss.NewStyle().Foreground(mutedColor).Render("· " + fmt.Sprintf(format, a...)))
}

func Success(format string, a ...any) {
	fmt.Println(box.BorderForeground(successColor).Foreground(successColor).Render("✔ " + fmt.Sprintf(format, a...)))
}

func Warn(format string, a ...any) {
	fmt.Println(box.BorderForeground(warnColor).Foreground(warnColor).Render("! " + fmt.Sprintf(format, a...)))
}

func ErrorBox(title, detail string) {
	body := lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render(title)
	if detail != "" {
		body += "\n" + detail
	}
	fmt.Fprintln(os.Stderr, box.BorderForeground(errorColor).Width(70).Render(body))
}

func Step(n, total int, title string) {
	line := lipgloss.NewStyle().Foreground(primary).Bold(true).Render(fmt.Sprintf("── %d/%d %s", n, total, title))
	fmt.Println("\n" + line)
}

func KeyPanel(name, pubkey string) string {
	t := lipgloss.NewStyle().Bold(true).Foreground(primary).Render(name)
	return box.BorderForeground(successColor).Width(66).Render(t + "\n" + pubkey)
}

func Pass(format string, a ...any) {
	fmt.Println(lipgloss.NewStyle().Foreground(successColor).Render("✔ " + fmt.Sprintf(format, a...)))
}

func Fail(name, detail string) {
	fmt.Println(lipgloss.NewStyle().Foreground(errorColor).Render("✘ "+name) +
		lipgloss.NewStyle().Foreground(mutedColor).Render(" — "+detail))
}
