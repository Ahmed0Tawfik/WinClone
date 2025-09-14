# WinClone - Windows Program Scanner

A simple command-line tool to scan and list all installed programs on Windows systems using Go packages for easy understanding.

## What it does

WinClone scans the Windows registry to find all installed programs and displays them in a clean, organized list. It shows:
- Program name
- Version number (if available)
- Installation path (if available)

## How to use

### Basic usage
```bash
# Show help
go run . --help

# Scan and list all programs (display on screen)
go run . scan

# Save results to JSON file
go run . scan --output programs.json

# Save results to text file
go run . scan --output programs.txt
```

### Building for global use
```bash
# Build the executable
go build -o winclone.exe

# Now you can use it from anywhere
./winclone.exe scan                    # Display on screen
./winclone.exe scan -o programs.json   # Save as JSON
./winclone.exe scan -o programs.txt   # Save as text
```

## Why This Approach is Better for Learning

### Using Go Packages vs Raw Windows API

**Before (Raw Windows API):**
- Complex Windows API calls with `syscall` package
- Manual UTF-16 string conversion
- Manual memory management
- Hard to understand error handling
- Lots of low-level code

**Now (Go Packages):**
- Simple `golang.org/x/sys/windows/registry` package
- Automatic string conversion
- Automatic memory management
- Clear error handling
- High-level, readable code

## How it works (Step by Step)

### 1. Command Structure (Cobra Framework)
```go
// cmd/root.go - Sets up the main command
var rootCmd = &cobra.Command{
    Use:   "winclone",
    Short: "Windows Program Scanner",
    // ... more configuration
}

// cmd/scan.go - Defines the scan command
var scanCmd = &cobra.Command{
    Use: "scan",
    Run: func(cmd *cobra.Command, args []string) {
        // This runs when user types "winclone scan"
    },
}
```

**What Cobra does for us:**
- Parses command line arguments automatically
- Shows help messages
- Handles command routing
- Provides consistent CLI experience

### 2. Registry Scanning Process

#### Step 1: Direct Function Call
```go
programs, err := scanAllPrograms()
```
- Calls the scanning function directly (no structs needed!)

#### Step 2: Scan Both Locations
```go
// 64-bit programs
programs64, err := scanRegistryLocation(`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`)

// 32-bit programs  
programs32, err := scanRegistryLocation(`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`)
```

#### Step 3: Open Registry Key
```go
key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
```
**What this does:**
- Opens the registry key for reading
- Handles all Windows API complexity automatically
- Returns a Go-friendly key object

#### Step 4: Read Subkey Names
```go
subkeyNames, err := key.ReadSubKeyNames(-1)
```
**What this does:**
- Gets all subkey names (each subkey = one program)
- `-1` means "read all subkeys"
- No manual enumeration needed!

#### Step 5: Process Each Program
```go
for _, subkeyName := range subkeyNames {
    program, err := scanner.getProgramFromSubkey(key, subkeyName)
    // ... add to results
}
```

#### Step 6: Extract Program Info
```go
// Open the program's subkey
subkey, err := registry.OpenKey(parentKey, subkeyName, registry.QUERY_VALUE)

// Read program name
name, _, err := subkey.GetStringValue("DisplayName")

// Read version (optional)
version, _, err := subkey.GetStringValue("DisplayVersion")

// Read path (optional)  
path, _, err := subkey.GetStringValue("InstallLocation")
```

**What this does:**
- Opens each program's registry entry
- Reads the three key values we need
- Handles missing values gracefully (some programs don't have all fields)

### 3. Key Functions Explained

#### `scanAllPrograms()`
- **Purpose**: Main coordinator function
- **How it works**: Calls `scanRegistryLocation()` for both 64-bit and 32-bit locations
- **Why it's simple**: Just two function calls, no structs or complex logic

#### `scanRegistryLocation(keyPath string)`
- **Purpose**: Scans one registry location (64-bit or 32-bit)
- **How it works**: 
  1. Open registry key (one line!)
  2. Read all subkey names (one line!)
  3. Loop through each subkey
  4. Extract program info from each subkey
- **Why it's simple**: Registry package handles all the complexity

#### `getProgramFromSubkey(parentKey, subkeyName)`
- **Purpose**: Extracts program details from one registry subkey
- **How it works**:
  1. Open the subkey (one line!)
  2. Read DisplayName (one line!)
  3. Read DisplayVersion (one line, ignore errors)
  4. Read InstallLocation (one line, ignore errors)
- **Why it's simple**: Each registry read is just one function call

## Project Structure

```
WinClone/
├── main.go          # Entry point - just calls cmd.Execute()
├── cmd/
│   ├── root.go      # Cobra root command setup
│   └── scan.go      # Scan command and registry logic
├── go.mod           # Dependencies: cobra + registry package
└── README.md        # This file
```

## Learning Benefits

### 1. **Package Usage**
- Learn how to use Go packages effectively
- Understand dependency management with `go.mod`
- See how packages abstract complexity

### 2. **CLI Framework (Cobra)**
- Learn modern CLI development patterns
- Understand command structure and routing
- See how frameworks simplify development

### 3. **Registry Package**
- Learn Windows-specific Go packages
- Understand how Go wraps system APIs
- See clean error handling patterns

### 4. **Code Organization**
- Learn proper project structure
- Understand separation of concerns
- See how to write maintainable code

## Comparison: Before vs After

| Aspect | Raw Windows API | Go Packages |
|--------|----------------|-------------|
| **Lines of code** | ~200 lines | ~100 lines |
| **Complexity** | High (manual API calls) | Low (package methods) |
| **Error handling** | Manual error codes | Go error interface |
| **String handling** | Manual UTF-16 conversion | Automatic |
| **Memory management** | Manual handle closing | Automatic with defer |
| **Readability** | Hard to understand | Easy to follow |
| **Maintainability** | Difficult | Easy |

## Future Enhancements

- Export results to JSON/CSV
- Filter programs by name or type
- Compare program lists between systems
- Generate installation scripts
- Add progress bars for large scans
- Cache results for faster subsequent scans