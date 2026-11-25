package tui

import "charm.land/lipgloss/v2"

const welcomeText = `                                    
                                    
⠀⠀⠀⠀⣀⣤⣴⣶⣶⣦⣄⡀⠀⠀⠀⠀⠀⠀⢀⣤⣶⣶⣶⣶⣤⣀⠀⠀⠀⠀
⠀⠀⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀⠀⠀⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀⠀⠀
⠀⠀⣾⣿⠟⠋⠉⠀⠀⠉⠙⢿⣿⣿⡄⠀⢠⣿⣿⡿⠋⠉⠀⠀ ⠉⠻⣿⣷⠀⠀
⠀⢸⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠙⣿⣿⣿⣿⣿⠏⠀⠀⠀⠀⠀⠀⠀ ⠸⣿⡇⠀
⠀⢸⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢈⣿⣿⣿⣁⠀⠀⠀⠀⠀⠀⠀⠀⠀ ⣿⡇⠀
⠀⢸⣿⡆⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⣿⣿⣿⣿⣄⠀⠀⠀⠀⠀⠀⠀ ⢰⣿⡇⠀
⠀⠀⢿⣿⣄⡀⠀⠀⠀⢀⣀⣴⣿⣿⡿⠁⠻⣿⣿⣦⣀⠀⠀⠀⢀⣀⣴⣿⡿⠀⠀
⠀⠀⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⡿⠁⠀⠀
⠀⠀⠀⠀⠈⠛⠻⠿⠿⠿⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠿⠿⠿⠿⠛⠁⠀⠀⠀
                                    
        Momentum Inc.               
                 
	The truth is out there\nAgent Scully
`

// const welcomeText = "The truth is out there\nAgent Scully"

func renderWelcomeArea(width, height int) string {
	boxStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Background(baseBgColorDark).
		Foreground(baseFgColor).
		Padding(1, 3)

	if width > 10 {
		boxStyle = boxStyle.MaxWidth(width - 4)
	}

	box := boxStyle.Render(welcomeText)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}
