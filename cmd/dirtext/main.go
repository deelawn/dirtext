package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Get current directory
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Load gitignore patterns
	ignorePatterns, err := loadGitignore(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: couldn't load .gitignore: %v\n", err)
	}

	// Print the root directory name
	fmt.Println(filepath.Base(rootDir))

	// Walk the directory tree
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == rootDir {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// Skip hidden files and directories (starting with .)
		if isHidden(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files/directories that match gitignore patterns
		if shouldIgnore(relPath, d.IsDir(), ignorePatterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate the depth to determine indentation
		depth := strings.Count(relPath, string(os.PathSeparator))

		// Print the tree branch and the file/directory name
		fmt.Printf("%s%s%s\n",
			strings.Repeat("│   ", depth),
			"├── ",
			filepath.Base(path))

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

// isHidden checks if a file or directory is hidden (starts with .)
func isHidden(path string) bool {
	// Split the path into components
	parts := strings.Split(path, string(os.PathSeparator))

	// Check if any component starts with a dot
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}

	return false
}

// loadGitignore loads patterns from .gitignore file
func loadGitignore(rootDir string) ([]string, error) {
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	file, err := os.Open(gitignorePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Clean up the pattern
		pattern := line

		// Remove leading slashes for relative patterns
		pattern = strings.TrimPrefix(pattern, "/")

		// Handle directory-only patterns (ending with /)
		pattern = strings.TrimSuffix(pattern, "/")

		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// shouldIgnore checks if a path should be ignored based on gitignore patterns
func shouldIgnore(path string, isDir bool, patterns []string) bool {
	for _, pattern := range patterns {
		// Handle negated patterns
		if strings.HasPrefix(pattern, "!") {
			negatedPattern := strings.TrimPrefix(pattern, "!")
			if match(path, negatedPattern, isDir) {
				return false
			}
			continue
		}

		// Check if the pattern matches
		if match(path, pattern, isDir) {
			return true
		}
	}

	return false
}

// match checks if a path matches a gitignore pattern
func match(path string, pattern string, isDir bool) bool {
	// Convert gitignore glob pattern to Go's filepath.Match pattern
	// This is a simplified implementation

	// Handle directory wildcards (**)
	if strings.Contains(pattern, "**") {
		// Replace ** with a special marker
		pattern = strings.Replace(pattern, "**", "[[RECURSIVE]]", -1)

		// Split both the pattern and path into components
		patternParts := strings.Split(pattern, string(os.PathSeparator))
		pathParts := strings.Split(path, string(os.PathSeparator))

		return recursiveMatch(pathParts, patternParts, 0, 0)
	}

	// Simple matching using filepath.Match
	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}

	// Check for partial path match
	// e.g., if pattern is "build", it should match both "build" and "path/to/build"
	return strings.HasSuffix(path, pattern) ||
		strings.Contains(path, pattern+string(os.PathSeparator))
}

// recursiveMatch handles ** pattern matching
func recursiveMatch(path, pattern []string, pathIdx, patternIdx int) bool {
	// End conditions
	if patternIdx >= len(pattern) {
		return pathIdx >= len(path)
	}

	if pathIdx >= len(path) {
		// Check if remaining patterns are all **
		for i := patternIdx; i < len(pattern); i++ {
			if pattern[i] != "[[RECURSIVE]]" {
				return false
			}
		}
		return true
	}

	// Handle ** pattern
	if pattern[patternIdx] == "[[RECURSIVE]]" {
		// Try to match at current position or skip this path component
		return recursiveMatch(path, pattern, pathIdx+1, patternIdx) ||
			recursiveMatch(path, pattern, pathIdx, patternIdx+1) ||
			recursiveMatch(path, pattern, pathIdx+1, patternIdx+1)
	}

	// Regular pattern matching
	match, _ := filepath.Match(pattern[patternIdx], path[pathIdx])
	if match {
		return recursiveMatch(path, pattern, pathIdx+1, patternIdx+1)
	}

	return false
}
