package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/registry"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan and list all installed programs",
	Long: `Scan the Windows registry to find all installed programs.

This command will:
1. Open the Windows registry
2. Look in the Uninstall keys for both 64-bit and 32-bit programs
3. Extract program names, versions, and installation paths
4. Display the results in a clean format

The registry locations scanned:
- SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall (64-bit programs)
- SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall (32-bit programs)

Output Options:
- Display on screen (default): Shows programs in a numbered list
- JSON file (.json): Saves structured data for programming/APIs
- Text file (.txt): Saves human-readable format for documentation

Examples:
  winclone scan                    # Display on screen
  winclone scan -o programs.json   # Save as JSON
  winclone scan -o programs.txt    # Save as text file`,
	Run: func(cmd *cobra.Command, args []string) {
		// This function runs when the user types "winclone scan"
		fmt.Println("WinClone - Scanning installed programs...")
		fmt.Println("==========================================")

		// Run the scan directly - no need for a scanner struct!
		programs, err := scanAllPrograms()
		if err != nil {
			fmt.Printf("Error scanning programs: %v\n", err)
			return
		}

		// Check if user wants file output
		outputFile, _ := cmd.Flags().GetString("output")
		if outputFile != "" {
			// Determine format based on file extension
			if strings.HasSuffix(strings.ToLower(outputFile), ".json") {
				// Save to JSON file
				err := saveToJSON(programs, outputFile)
				if err != nil {
					fmt.Printf("Error saving to JSON: %v\n", err)
					return
				}
				fmt.Printf("\nResults saved to JSON: %s\n", outputFile)
			} else {
				// Save to text file
				err := saveToText(programs, outputFile)
				if err != nil {
					fmt.Printf("Error saving to text: %v\n", err)
					return
				}
				fmt.Printf("\nResults saved to text: %s\n", outputFile)
			}
		} else {
			// Display the results on screen
			displayResults(programs)
		}
	},
}

// Program represents an installed application
type Program struct {
	Name    string // Display name of the program
	Version string // Version number
	Path    string // Installation path
}

// scanAllPrograms scans both 64-bit and 32-bit program locations
// This is the main function that coordinates the entire scanning process
func scanAllPrograms() ([]Program, error) {
	var allPrograms []Program

	// Step 1: Scan 64-bit programs
	fmt.Println("Step 1: Scanning 64-bit programs...")
	fmt.Println("Location: SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall")

	programs64, err := scanRegistryLocation(`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`)
	if err != nil {
		fmt.Printf("Warning: Could not scan 64-bit programs: %v\n", err)
	} else {
		fmt.Printf("Found %d 64-bit programs\n", len(programs64))
		allPrograms = append(allPrograms, programs64...)
	}

	// Step 2: Scan 32-bit programs (WOW64 = Windows on Windows 64-bit)
	fmt.Println("\nStep 2: Scanning 32-bit programs...")
	fmt.Println("Location: SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall")

	programs32, err := scanRegistryLocation(`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`)
	if err != nil {
		fmt.Printf("Warning: Could not scan 32-bit programs: %v\n", err)
	} else {
		fmt.Printf("Found %d 32-bit programs\n", len(programs32))
		allPrograms = append(allPrograms, programs32...)
	}

	return allPrograms, nil
}

// scanRegistryLocation opens a registry key and scans all its subkeys
// Each subkey represents one installed program
func scanRegistryLocation(keyPath string) ([]Program, error) {
	var programs []Program

	// Step 1: Open the registry key
	// registry.OpenKey() is much simpler than raw Windows API calls!
	// It handles all the UTF-16 conversion and error handling for us
	fmt.Printf("  Opening registry key: %s\n", keyPath)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %v", err)
	}
	defer key.Close() // Always close the key when done

	// Step 2: Get all subkey names
	// registry.ReadSubKeyNames() does all the enumeration work for us
	fmt.Printf("  Reading subkey names...\n")
	subkeyNames, err := key.ReadSubKeyNames(-1) // -1 means read all subkeys
	if err != nil {
		return nil, fmt.Errorf("failed to read subkey names: %v", err)
	}

	fmt.Printf("  Found %d subkeys to process\n", len(subkeyNames))

	// Step 3: Process each subkey (each subkey = one program)
	for i, subkeyName := range subkeyNames {
		// Show progress every 50 programs
		if i%50 == 0 && i > 0 {
			fmt.Printf("  Processed %d/%d programs...\n", i, len(subkeyNames))
		}

		// Get program info from this subkey
		program, err := getProgramFromSubkey(key, subkeyName)
		if err != nil {
			// Skip programs that can't be read (some are system components)
			continue
		}

		// Only add programs that have a name (some entries are just metadata)
		if program.Name != "" {
			programs = append(programs, program)
		}
	}

	return programs, nil
}

// getProgramFromSubkey reads program details from a specific registry subkey
// This function extracts the DisplayName, DisplayVersion, and InstallLocation
func getProgramFromSubkey(parentKey registry.Key, subkeyName string) (Program, error) {
	var program Program

	// Step 1: Open the subkey
	// This opens the specific program's registry entry
	subkey, err := registry.OpenKey(parentKey, subkeyName, registry.QUERY_VALUE)
	if err != nil {
		return program, err
	}
	defer subkey.Close()

	// Step 2: Read the DisplayName
	// This is the name you see in "Add or Remove Programs"
	name, _, err := subkey.GetStringValue("DisplayName")
	if err != nil {
		// Some programs don't have a DisplayName, skip them
		return program, err
	}
	program.Name = strings.TrimSpace(name) // Remove extra whitespace

	// Step 3: Read the DisplayVersion (optional)
	// Not all programs have this, so we ignore errors
	version, _, err := subkey.GetStringValue("DisplayVersion")
	if err == nil {
		program.Version = strings.TrimSpace(version)
	}

	// Step 4: Read the InstallLocation (optional)
	// This is where the program is installed
	path, _, err := subkey.GetStringValue("InstallLocation")
	if err == nil {
		program.Path = strings.TrimSpace(path)
	}

	return program, nil
}

// displayResults formats and displays the scan results
func displayResults(programs []Program) {
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("SCAN COMPLETE!\n")
	fmt.Printf("Found %d installed programs:\n", len(programs))
	fmt.Printf(strings.Repeat("=", 50) + "\n\n")

	// Display each program with nice formatting
	for i, program := range programs {
		fmt.Printf("%d. %s", i+1, program.Name)

		// Add version if available
		if program.Version != "" {
			fmt.Printf(" (v%s)", program.Version)
		}
		fmt.Println()

		// Add installation path if available
		if program.Path != "" {
			fmt.Printf("   Path: %s\n", program.Path)
		}
		fmt.Println()
	}
}

// saveToJSON saves the program list to a JSON file
func saveToJSON(programs []Program, filename string) error {
	// Create the JSON file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create JSON encoder
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print with 2-space indentation

	// Encode the programs slice to JSON
	err = encoder.Encode(programs)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	return nil
}

// saveToText saves the program list to a text file
func saveToText(programs []Program, filename string) error {
	// Create the text file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "WinClone - Installed Programs List\n")
	fmt.Fprintf(file, "Generated on: %s\n", "2025-01-14") // You could use time.Now() here
	fmt.Fprintf(file, "Total programs found: %d\n", len(programs))
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 50))

	// Write each program
	for i, program := range programs {
		fmt.Fprintf(file, "%d. %s", i+1, program.Name)

		// Add version if available
		if program.Version != "" {
			fmt.Fprintf(file, " (v%s)", program.Version)
		}
		fmt.Fprintf(file, "\n")

		// Add installation path if available
		if program.Path != "" {
			fmt.Fprintf(file, "   Path: %s\n", program.Path)
		}
		fmt.Fprintf(file, "\n")
	}

	return nil
}

func init() {
	// This function runs when the package is initialized
	// It adds the scan command to the root command and sets up flags
	rootCmd.AddCommand(scanCmd)

	// Add the --output flag for file export
	scanCmd.Flags().StringP("output", "o", "", "Save results to file (JSON: .json, Text: .txt)")
}
