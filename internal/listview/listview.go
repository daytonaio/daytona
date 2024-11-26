package listview

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	helpToggleKey = "?"
	quitKey       = "q"
)

// ListView represents a list view with items and help visibility.
type ListView struct {
	items       []string
	helpVisible bool
}

// New creates and returns a new ListView instance.
func New() *ListView {
	return &ListView{
		items:       []string{},
		helpVisible: false,
	}
}

// Render displays the list view and handles user input.
func (lv *ListView) Render() {
	for {
		lv.clearScreen()
		lv.displayItems()
		if lv.helpVisible {
			lv.displayHelp()
		}
		lv.displayFooter()

		input := lv.getInput()
		if input == quitKey {
			break
		} else if input == helpToggleKey {
			lv.toggleHelp()
		}
	}
}

func (lv *ListView) clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error clearing screen:", err)
	}
}

func (lv *ListView) displayItems() {
	for _, item := range lv.items {
		fmt.Println(item)
	}
}

func (lv *ListView) displayHelp() {
	fmt.Println("\nHelp Menu:")
	fmt.Printf("%s - Toggle help\n", helpToggleKey)
	fmt.Printf("%s - Quit\n", quitKey)
}

func (lv *ListView) displayFooter() {
	fmt.Printf("\nPress '%s' for help, '%s' to quit\n", helpToggleKey, quitKey)
}

func (lv *ListView) getInput() string {
	var input string
	fmt.Scanln(&input)
	return input
}

func (lv *ListView) toggleHelp() {
	lv.helpVisible = !lv.helpVisible
}
