package main

import (
	"fmt"
  "bufio"
  "io"
  "os/user"
  "encoding/json"
  "path/filepath"
	"os"
  "strings"

	"github.com/fatih/color"
)

var newDir string

var pathsToTrack = []PathInfo{
	{"dwm", "config.h"},
	{"dwmblocks", "blocks.h"},
}

// Show usage instructions
type DirectoryInfo struct {
	DirectoryName string    `json:"directoryName"`
	DirectoryPath string    `json:"directoryPath"`
	Files         []FileInfo `json:"files"`
}

type FileInfo struct {
	OriginalPath string `json:"originalPath"`
	NewPath      string `json:"newPath"`
}

type PathInfo struct {
	DirName   string
	FileName  string
}


func showUsage() {
	color.Magenta("╔═══════════════════════════════════════════╗")
	color.Magenta("║      (っ◔◡◔)っ ♥ DOTFILE SAVER ♥          ║")
	color.Magenta("╚═══════════════════════════════════════════╝")
	fmt.Println("Made by Azsao")
	color.Yellow("something")
	color.Yellow("something 2")
	masterQuestion()
}


func masterQuestion() {
	reader := bufio.NewReader(os.Stdin)

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		return
	}

	homeDir := usr.HomeDir
	dotSaverDir := filepath.Join(homeDir, "dot-saver")

	if _, err := os.Stat(dotSaverDir); os.IsNotExist(err) {
		err := os.Mkdir(dotSaverDir, 0755)
		if err != nil {
			fmt.Printf("Error creating dot-saver directory: %v\n", err)
			return
		}
	}

	color.Cyan("Enter the name of the directory you would like to create: ")
	dirName, _ := reader.ReadString('\n')
	dirName = trimNewline(dirName)

	newDir = filepath.Join(dotSaverDir, dirName)
	err = os.Mkdir(newDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	color.Green(fmt.Sprintf("Directory '%s' created successfully in '%s'.\n", dirName, dotSaverDir))

	// Create initial DirectoryInfo struct with no files
	dirInfo := DirectoryInfo{
		DirectoryName: dirName,
		DirectoryPath: newDir,
		Files:         []FileInfo{},
	}

	jsonData, err := json.MarshalIndent(dirInfo, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	jsonFilePath := filepath.Join(newDir, "directory.json")
	err = os.WriteFile(jsonFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON file: %v\n", err)
		return
	}

	color.Green(fmt.Sprintf("JSON file '%s' created successfully.\n", jsonFilePath))

	slaveQuestion()
}

func expandHomeDir(path string) (string, error) {
    if strings.HasPrefix(path, "~/") {
        usr, err := user.Current()
        if err != nil {
            return "", err
        }
        return filepath.Join(usr.HomeDir, path[2:]), nil
    } else if path == "~" {
        usr, err := user.Current()
        if err != nil {
            return "", err
        }
        return usr.HomeDir, nil
    }
    return path, nil
}


func slaveQuestion() {
    reader := bufio.NewReader(os.Stdin)

    // Ask user if they want to copy a custom path
    color.Red("Would you like to copy a custom path? (yes/no): ")
    answer, _ := reader.ReadString('\n')
    answer = trimNewline(answer)

    if answer == "yes" {
        // Ask for custom paths
        color.Red("Enter the path(s) separated by commas (e.g., ~/path/to/file1, ~/path/to/dir2): ")
        pathsInput, _ := reader.ReadString('\n')
        pathsInput = trimNewline(pathsInput)

        // Split paths by comma
        paths := strings.Split(pathsInput, ",")

        // Validate each path
        for _, path := range paths {
            path = strings.TrimSpace(path)
            // Expand ~ to user's home directory
            expandedPath, err := expandHomeDir(path)
            if err != nil {
                fmt.Printf("Error expanding home directory in path '%s': %v\n", path, err)
                return
            }
            if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
                fmt.Printf("Path '%s' does not exist.\n", path)
                return
            }
        }

        // Copy each valid path into the newly created directory
        for _, path := range paths {
            path = strings.TrimSpace(path)
            // Expand ~ to user's home directory
            expandedPath, err := expandHomeDir(path)
            if err != nil {
                fmt.Printf("Error expanding home directory in path '%s': %v\n", path, err)
                return
            }
            base := filepath.Base(expandedPath)
            destPath := filepath.Join(newDir, base)

            err = copyFileOrDir(expandedPath, destPath)
            if err != nil {
                fmt.Printf("Error copying '%s' to '%s': %v\n", expandedPath, destPath, err)
                return
            }

            color.Green(fmt.Sprintf("Successfully copied '%s' to '%s'.\n", expandedPath, destPath))
        }
    } else if answer == "no" {
        fmt.Println("Write 'skip' to skip.")
    } else {
        fmt.Println("Invalid input. Write 'skip' to skip.")
    }
}


func originalUse() {
	// Get current user info
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		return
	}
	homeDir := usr.HomeDir

	// Define the savedDotfile directory
	savedDotfileDir := filepath.Join(homeDir, "savedDotfile")

	// Initialize DirectoryInfo struct
	dirInfo := DirectoryInfo{
		DirectoryName: filepath.Base(newDir),
		DirectoryPath: newDir,
		Files:         []FileInfo{},
	}

	// Walk through the home directory, ignoring savedDotfile directory
	err = filepath.Walk(homeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the savedDotfile directory
		if info.IsDir() && path == savedDotfileDir {
			return filepath.SkipDir
		}

		// Process directories and files
		if info.IsDir() {
			dirName := filepath.Base(path)
			for _, pathInfo := range pathsToTrack {
				if dirName == pathInfo.DirName {
					filePath := filepath.Join(path, pathInfo.FileName)
					if fileExists(filePath) {
						destDir := filepath.Join(newDir, pathInfo.DirName)
						destFilePath := filepath.Join(destDir, pathInfo.FileName)

						err := os.MkdirAll(destDir, 0755)
						if err != nil {
							fmt.Printf("Error creating directory '%s': %v\n", destDir, err)
							return nil
						}

						err = copyFile(filePath, destFilePath)
						if err != nil {
							fmt.Printf("Error copying '%s' to '%s': %v\n", filePath, destFilePath, err)
							return nil
						}

						// Append file info to directory info
						dirInfo.Files = append(dirInfo.Files, FileInfo{
							OriginalPath: filePath,
							NewPath:      destFilePath,
						})

						color.Green(fmt.Sprintf("Successfully copied '%s' to '%s'.\n", filePath, destFilePath))
					} else {
						color.Yellow(fmt.Sprintf("File '%s' does not exist in directory '%s'.\n", pathInfo.FileName, path))
					}
				}
			}
		}
		return nil
	})
	// Update the JSON file with the directory info
	jsonData, err := json.MarshalIndent(dirInfo, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	jsonFilePath := filepath.Join(newDir, "directory.json")
	err = os.WriteFile(jsonFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error updating JSON file: %v\n", err)
		return
	}
  endWrap()
}
func endWrap() {
  	// Get current user info
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		return
	}

	// Define paths
	homeDir := usr.HomeDir
	savedDotfileDir := filepath.Join(homeDir, "savedDotfile")

	// Check if ~/savedDotfile exists, create if not
	if _, err := os.Stat(savedDotfileDir); os.IsNotExist(err) {
		err := os.Mkdir(savedDotfileDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory '%s': %v\n", savedDotfileDir, err)
			return
		}
		color.Green(fmt.Sprintf("Directory '%s' created successfully.\n", savedDotfileDir))
	}

	// Define destination directory
	destDir := filepath.Join(savedDotfileDir, filepath.Base(newDir))

	// Move the directory
	err = moveDirectory(newDir, destDir)
	if err != nil {
		fmt.Printf("Error moving directory '%s' to '%s': %v\n", newDir, destDir, err)
		return
	}

	color.Green(fmt.Sprintf("Directory moved successfully to '%s'.\n", destDir))
}

// Helper function to move a directory
func moveDirectory(sourceDir, destDir string) error {
	// Move the directory
	err := os.Rename(sourceDir, destDir)
	if err != nil {
		return fmt.Errorf("error moving directory from '%s' to '%s': %v", sourceDir, destDir, err)
	}
	return nil
}

// trimNewline removes trailing newline characters from a string
func trimNewline(s string) string {
	return strings.TrimSuffix(s, "\n")   
}

// Check if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// Check if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// copyFileOrDir copies a file or directory from source to destination
func copyFileOrDir(source, dest string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		return copyDir(source, dest)
	} else {
		return copyFile(source, dest)
	}
}

// copyDir copies a directory from source to destination
func copyDir(source, dest string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, sourceInfo.Mode())
	if err != nil {
		return err
	}

	dir, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range dir {
		srcPath := filepath.Join(source, entry.Name())
		dstPath := filepath.Join(dest, entry.Name())

		err = copyFileOrDir(srcPath, dstPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a file from source to destination
func copyFile(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.Chmod(dest, sourceInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}




func main() {
	showUsage()
  originalUse()
}
