package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func PrintBanner(version string) {
	sparkle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E8915A")).
		Bold(true)

	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D4A574")).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555")).
		Italic(true)

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))

	art := `
  ╔═╗ ╦  ╔═╗ ╔╗    ╔═╗ ╦  ╔═╗ ╦ ╦ ╔═╗ ╦═╗ ╔═╗ ╦ ╦ ╔╗╔ ╔╦╗
  ╠═╝ ║  ╠═╣ ╠╩╗   ╠═╝ ║  ╠═╣ ╚╦╝ ║ ╦ ╠╦╝ ║ ║ ║ ║ ║║║  ║║
  ╩   ╩═╝╩ ╩ ╚═╝   ╩   ╩═╝╩ ╩  ╩  ╚═╝ ╩╚═ ╚═╝ ╚═╝ ╝╚╝ ═╩╝`

	fmt.Println()
	fmt.Printf("  %s\n", sparkle.Render("✻"))
	fmt.Println(logo.Render(art))
	fmt.Println()
	fmt.Printf("  %s  %s\n", subtitleStyle.Render("우리가 만들어내는 무한한 가능성"), versionStyle.Render(version))
	fmt.Println()
	fmt.Printf("  %s\n", divider.Render("────────────────────────────────"))
	fmt.Println()
	fmt.Printf("  %s\n", hintStyle.Render("종료하려면 Control + C 키를 함께 눌러주세요 (⌘ 아님 주의!)"))
	fmt.Println()
}
