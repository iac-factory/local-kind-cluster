Certainly! Below is the Go program using the `cobra` and `viper` packages to create a CLI tool with the specified commands and functionality. The program includes:

- Two parent commands: `scan` and `add`.
- Each parent command contains a `package-documentation` child command.
- Functionality to recursively walk through Go packages, check for `doc.go` files, and validate or create them according to Go's documentation standards.

```go
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gopkgdoc",
		Short: "Go Package Documentation CLI",
	}

	// Add parent commands
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(addCmd)

	// Add child commands
	scanCmd.AddCommand(scanPackageDocCmd)
	addCmd.AddCommand(addPackageDocCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Check if go.mod exists in the current directory
func checkGoMod() error {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return errors.New("go.mod not found in the current directory; please run this command from the root of a Go project")
	}
	return nil
}

// Check if a directory contains any Go files
func containsGoFiles(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
			return true
		}
	}
	return false
}

// Validate the doc.go file content
func validateDocFile(path string, packageName string) (bool, string) {
	docFile := filepath.Join(path, "doc.go")
	content, err := ioutil.ReadFile(docFile)
	if err != nil {
		return false, ""
	}

	// Regular expression to match Go's doc.go documentation standard
	re := regexp.MustCompile(`(?m)^// Package\s+` + regexp.QuoteMeta(packageName) + `\s+.*\npackage\s+` + regexp.QuoteMeta(packageName) + `\s*$`)

	if re.Match(content) {
		return true, strings.TrimSpace(string(content))
	}
	return false, strings.TrimSpace(string(content))
}

// SCAN COMMANDS

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan commands",
}

var scanPackageDocCmd = &cobra.Command{
	Use:   "package-documentation",
	Short: "Scan for Go packages with documentation",
	RunE:  scanPackageDocumentation,
}

func scanPackageDocumentation(cmd *cobra.Command, args []string) error {
	if err := checkGoMod(); err != nil {
		return err
	}

	type ReportItem struct {
		PackagePath   string
		PackageName   string
		RelativePath  string
		DocExists     bool
		DocValid      bool
		DocContent    string
		ErrorMessage  string
	}

	var report []ReportItem

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.Contains(path, "/.") {
			return nil
		}

		if info.IsDir() && containsGoFiles(path) {
			packageName := filepath.Base(path)
			docExists := false
			docValid := false
			docContent := ""
			errorMessage := ""

			if _, err := os.Stat(filepath.Join(path, "doc.go")); err == nil {
				docExists = true
				valid, content := validateDocFile(path, packageName)
				docValid = valid
				docContent = content
				if !valid {
					errorMessage = "doc.go does not meet Go documentation standards"
				}
			} else {
				errorMessage = "doc.go not found"
			}

			report = append(report, ReportItem{
				PackagePath:  path,
				PackageName:  packageName,
				RelativePath: path,
				DocExists:    docExists,
				DocValid:     docValid,
				DocContent:   docContent,
				ErrorMessage: errorMessage,
			})
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Generate report
	fmt.Println("Package Documentation Scan Report:")
	for _, item := range report {
		fmt.Printf("\nPackage: %s\n", item.PackageName)
		fmt.Printf("Relative Path: %s\n", item.RelativePath)
		if item.DocExists && item.DocValid {
			fmt.Println("Status: Passed")
			fmt.Printf("Documentation:\n%s\n", item.DocContent)
		} else {
			fmt.Println("Status: Failed")
			fmt.Printf("Reason: %s\n", item.ErrorMessage)
			if item.DocContent != "" {
				fmt.Printf("Current doc.go content:\n%s\n", item.DocContent)
			}
		}
	}
	return nil
}

// ADD COMMANDS

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add commands",
}

var addPackageDocCmd = &cobra.Command{
	Use:   "package-documentation",
	Short: "Add or update package documentation",
	RunE:  addPackageDocumentation,
}

func addPackageDocumentation(cmd *cobra.Command, args []string) error {
	if err := checkGoMod(); err != nil {
		return err
	}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.Contains(path, "/.") {
			return nil
		}

		if info.IsDir() && containsGoFiles(path) {
			packageName := filepath.Base(path)
			needsUpdate := false

			docFilePath := filepath.Join(path, "doc.go")
			if _, err := os.Stat(docFilePath); os.IsNotExist(err) {
				needsUpdate = true
			} else {
				valid, _ := validateDocFile(path, packageName)
				if !valid {
					needsUpdate = true
				}
			}

			if needsUpdate {
				docContent := fmt.Sprintf("// Package %s @TODO document the package.\npackage %s\n", packageName, packageName)
				err := ioutil.WriteFile(docFilePath, []byte(docContent), 0644)
				if err != nil {
					fmt.Printf("Failed to write doc.go for package '%s': %v\n", packageName, err)
				} else {
					fmt.Printf("doc.go created/updated for package '%s' at '%s'\n", packageName, path)
				}
			} else {
				fmt.Printf("doc.go already valid for package '%s' at '%s'\n", packageName, path)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}
```

### How to Use This Program

1. **Install Dependencies:**

   Make sure you have the `cobra` and `viper` packages installed:

   ```bash
   go get -u github.com/spf13/cobra
   go get -u github.com/spf13/viper
   ```

2. **Build the Program:**

   Save the code to a file named `main.go` and build it:

   ```bash
   go build -o gopkgdoc main.go
   ```

3. **Run the Scan Command:**

   ```bash
   ./gopkgdoc scan package-documentation
   ```

   This will generate a report of all packages and whether they have valid `doc.go` files.

4. **Run the Add Command:**

   ```bash
   ./gopkgdoc add package-documentation
   ```

   This will create or update `doc.go` files for packages that are missing them or have invalid documentation.

### Notes

- **Error Handling:** The program will fail immediately if `go.mod` is not found in the current working directory.
- **Package Identification:** It identifies a directory as a Go package if it contains any `.go` files.
- **Documentation Standards:** The program checks for `doc.go` files that start with `// Package [package-name]` followed by `package [package-name]`.
- **Metadata in Reports:** The scan report includes the package name, relative path, documentation status, and actual documentation content if available.
- **Creating/Updating `doc.go`:** For the add command, if a `doc.go` file doesn't exist or doesn't meet the standards, it creates or updates it with a template containing a TODO comment.

### Example Output

**Scan Command:**

```bash
./gopkgdoc scan package-documentation
```

```
Package Documentation Scan Report:

Package: utils
Relative Path: ./utils
Status: Failed
Reason: doc.go not found

Package: models
Relative Path: ./models
Status: Passed
Documentation:
// Package models provides data models for the application.
package models
```

**Add Command:**

```bash
./gopkgdoc add package-documentation
```

```
doc.go created/updated for package 'utils' at './utils'
doc.go already valid for package 'models' at './models'
```

---

Feel free to customize and enhance the program according to your specific needs!