package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "winclone",
	Short: "Windows Program Scanner - Scan and list installed programs",
	Long: `WinClone is a simple tool that scans the Windows registry to find 
all installed programs and displays them in a clean, organized list.

It shows:
- Program name
- Version number (if available)  
- Installation path (if available)

The tool scans both 64-bit and 32-bit programs from the Windows registry.`,
	Example: `winclone scan                    # Display programs on screen
winclone scan -o programs.json   # Save as JSON file
winclone scan -o programs.txt     # Save as text file
winclone help                     # Show this help`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
