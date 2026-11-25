package main

import (
	"agent_scully/tui"
	"agent_scully/utils"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	conf, err := utils.ReadConfigFile("./config/config.json")
	if err != nil {
		fmt.Println("Config file not found. Using the default config")
		conf = utils.CreateDefaultAgentConfig()
	}
	p := tea.NewProgram(tui.InitializeModel(&conf))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
