// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// isRipgrepAvailable checks if ripgrep is installed and available
func isRipgrepAvailable() bool {
	_, err := exec.LookPath("rg")
	return err == nil
}

func SearchFiles(c *gin.Context) {
	// Only handle POST requests with JSON content
	if c.Request.Method != "POST" {
		c.AbortWithError(http.StatusMethodNotAllowed, errors.New("only POST method is allowed"))
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		c.AbortWithError(http.StatusBadRequest, errors.New("content-type must be application/json"))
		return
	}

	// Parse JSON request
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate required fields
	if req.Query == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("query is required"))
		return
	}

	// Set default path if not provided
	if req.Path == "" {
		req.Path = "."
	}

	// Immediately check ripgrep availability and decide on implementation
	if isRipgrepAvailable() {
		handleRipgrepSearch(c, req)
	} else {
		handleFallbackContentSearch(c, req)
	}
}

// handleRipgrepSearch uses ripgrep for content search
func handleRipgrepSearch(c *gin.Context, req SearchRequest) {
	// Validate path exists
	if !filepath.IsAbs(req.Path) {
		// Convert relative path to absolute
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid path: %w", err))
			return
		}
		req.Path = absPath
	}

	// Build ripgrep command
	args, err := buildRipgrepCommand(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to build command: %w", err))
		return
	}

	// Execute ripgrep command
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()

	// ripgrep returns exit code 1 when no matches are found, which is not an error
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means no matches found, which is valid
			if exitError.ExitCode() != 1 {
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("ripgrep command failed: %w", err))
				return
			}
		} else {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to execute ripgrep: %w", err))
			return
		}
	}

	// Parse the output
	results, err := parseRipgrepOutput(string(output), req)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to parse output: %w", err))
		return
	}

	c.JSON(http.StatusOK, results)
}

// handleFallbackContentSearch provides basic content search when ripgrep is not available
func handleFallbackContentSearch(c *gin.Context, req SearchRequest) {
	matches := []SearchMatch{}
	fileSet := make(map[string]bool)
	totalMatches := 0

	err := filepath.Walk(req.Path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		// Apply file type filtering if specified
		if len(req.FileTypes) > 0 {
			ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
			found := false
			for _, fileType := range req.FileTypes {
				if ext == fileType {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// Apply include/exclude globs if specified
		if len(req.IncludeGlobs) > 0 {
			included := false
			for _, glob := range req.IncludeGlobs {
				if matched, _ := filepath.Match(glob, filepath.Base(filePath)); matched {
					included = true
					break
				}
			}
			if !included {
				return nil
			}
		}

		for _, glob := range req.ExcludeGlobs {
			if matched, _ := filepath.Match(glob, filepath.Base(filePath)); matched {
				return nil
			}
		}

		file, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer file.Close()

		// Check if file is binary
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil {
			return nil
		}

		for i := 0; i < n; i++ {
			if buf[i] == 0 {
				return nil // Skip binary files
			}
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return nil
		}

		scanner := bufio.NewScanner(file)
		lineNum := 1
		fileMatches := 0

		for scanner.Scan() {
			line := scanner.Text()

			// Apply case sensitivity
			searchQuery := req.Query
			searchLine := line
			if req.CaseSensitive != nil && !*req.CaseSensitive {
				searchQuery = strings.ToLower(searchQuery)
				searchLine = strings.ToLower(line)
			}

			if strings.Contains(searchLine, searchQuery) {
				// Handle count-only mode
				if req.CountOnly != nil && *req.CountOnly {
					fileMatches++
					totalMatches++
				} else if req.FilenamesOnly != nil && *req.FilenamesOnly {
					fileSet[filePath] = true
					break // Only need to know the file contains matches
				} else {
					// Find column position
					column := strings.Index(searchLine, searchQuery) + 1

					match := SearchMatch{
						File:       filePath,
						LineNumber: lineNum,
						Column:     column,
						Line:       line,
						Match:      req.Query,
					}
					matches = append(matches, match)
					totalMatches++
					fileSet[filePath] = true

					// Apply max results limit
					if req.MaxResults != nil && totalMatches >= *req.MaxResults {
						return filepath.SkipAll
					}
				}
			}
			lineNum++
		}

		// For count-only mode, we still track the file if it had matches
		if fileMatches > 0 && (req.CountOnly != nil && *req.CountOnly) {
			fileSet[filePath] = true
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Build results
	results := &SearchResults{
		Matches:      matches,
		TotalMatches: totalMatches,
		TotalFiles:   len(fileSet),
		Files:        []string{},
	}

	// Populate files list
	for file := range fileSet {
		results.Files = append(results.Files, file)
	}

	c.JSON(http.StatusOK, results)
}

// RipgrepOutput represents the JSON output structure from ripgrep
type RipgrepOutput struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path,omitempty"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines,omitempty"`
		LineNumber     *int `json:"line_number,omitempty"`
		AbsoluteOffset *int `json:"absolute_offset,omitempty"`
		Submatches     []struct {
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
			Start int `json:"start"`
			End   int `json:"end"`
		} `json:"submatches,omitempty"`
	} `json:"data"`
}

// buildRipgrepCommand constructs the ripgrep command with all the specified options
func buildRipgrepCommand(req SearchRequest) ([]string, error) {
	if req.Query == "" {
		return nil, errors.New("query is required")
	}

	args := []string{"rg"}

	// Set path (default to current directory)
	searchPath := req.Path
	if searchPath == "" {
		searchPath = "."
	}

	// Case sensitivity
	if req.CaseSensitive != nil && !*req.CaseSensitive {
		args = append(args, "--ignore-case")
	}

	// Multiline mode
	if req.Multiline != nil && *req.Multiline {
		args = append(args, "--multiline")
	}

	// Context lines
	if req.Context != nil && *req.Context > 0 {
		args = append(args, "--context", strconv.Itoa(*req.Context))
	}

	// Count only
	if req.CountOnly != nil && *req.CountOnly {
		args = append(args, "--count")
	}

	// Filenames only
	if req.FilenamesOnly != nil && *req.FilenamesOnly {
		args = append(args, "--files-with-matches")
	}

	// JSON output (for structured parsing)
	if req.JSON != nil && *req.JSON {
		args = append(args, "--json")
	}

	// Max results
	if req.MaxResults != nil && *req.MaxResults > 0 {
		args = append(args, "--max-count", strconv.Itoa(*req.MaxResults))
	}

	// File types
	for _, fileType := range req.FileTypes {
		args = append(args, "--type", fileType)
	}

	// Include globs
	for _, glob := range req.IncludeGlobs {
		args = append(args, "--glob", glob)
	}

	// Exclude globs
	for _, glob := range req.ExcludeGlobs {
		args = append(args, "--glob", "!"+glob)
	}

	// Additional ripgrep arguments
	if len(req.RgArgs) > 0 {
		args = append(args, req.RgArgs...)
	}

	// Add line numbers and column numbers for better output parsing
	if req.JSON == nil || !*req.JSON {
		args = append(args, "--line-number", "--column")
	}

	// Add the query and path
	args = append(args, req.Query, searchPath)

	return args, nil
}

// parseRipgrepOutput parses the output from ripgrep and converts it to SearchResults
func parseRipgrepOutput(output string, req SearchRequest) (*SearchResults, error) {
	results := &SearchResults{
		Matches:      []SearchMatch{},
		TotalMatches: 0,
		TotalFiles:   0,
		Files:        []string{},
	}

	if req.JSON != nil && *req.JSON {
		return parseJSONOutput(output, results)
	}

	if req.CountOnly != nil && *req.CountOnly {
		return parseCountOutput(output, results)
	}

	if req.FilenamesOnly != nil && *req.FilenamesOnly {
		return parseFilenamesOutput(output, results)
	}

	return parseStandardOutput(output, results, req)
}

// parseJSONOutput parses JSON output from ripgrep
func parseJSONOutput(output string, results *SearchResults) (*SearchResults, error) {
	results.RawOutput = output
	scanner := bufio.NewScanner(strings.NewReader(output))
	fileSet := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var rgOutput RipgrepOutput
		if err := json.Unmarshal([]byte(line), &rgOutput); err != nil {
			continue // Skip invalid JSON lines
		}

		if rgOutput.Type == "match" {
			filePath := rgOutput.Data.Path.Text
			lineText := rgOutput.Data.Lines.Text
			lineNumber := 1
			if rgOutput.Data.LineNumber != nil {
				lineNumber = *rgOutput.Data.LineNumber
			}

			fileSet[filePath] = true

			// Extract matches from submatches
			for _, submatch := range rgOutput.Data.Submatches {
				match := SearchMatch{
					File:       filePath,
					LineNumber: lineNumber,
					Column:     submatch.Start + 1, // Convert to 1-based indexing
					Line:       lineText,
					Match:      submatch.Match.Text,
				}
				results.Matches = append(results.Matches, match)
				results.TotalMatches++
			}
		}
	}

	results.TotalFiles = len(fileSet)
	for file := range fileSet {
		results.Files = append(results.Files, file)
	}

	return results, nil
}

// parseCountOutput parses count-only output from ripgrep
func parseCountOutput(output string, results *SearchResults) (*SearchResults, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	fileSet := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Format: filename:count
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			filename := parts[0]
			countStr := parts[1]

			if count, err := strconv.Atoi(countStr); err == nil {
				fileSet[filename] = true
				results.TotalMatches += count
			}
		}
	}

	results.TotalFiles = len(fileSet)
	for file := range fileSet {
		results.Files = append(results.Files, file)
	}

	return results, nil
}

// parseFilenamesOutput parses filenames-only output from ripgrep
func parseFilenamesOutput(output string, results *SearchResults) (*SearchResults, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	fileSet := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			fileSet[line] = true
		}
	}

	results.TotalFiles = len(fileSet)
	for file := range fileSet {
		results.Files = append(results.Files, file)
	}

	return results, nil
}

// parseStandardOutput parses standard ripgrep output with line numbers and columns
func parseStandardOutput(output string, results *SearchResults, req SearchRequest) (*SearchResults, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	fileSet := make(map[string]bool)

	// Regex to parse ripgrep output: filename:line:column:content
	lineRegex := regexp.MustCompile(`^([^:]+):(\d+):(\d+):(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			filename := matches[1]
			lineNum, _ := strconv.Atoi(matches[2])
			column, _ := strconv.Atoi(matches[3])
			content := matches[4]

			fileSet[filename] = true

			// Extract the actual match from the content
			// For now, we'll use the query as the match (could be improved with regex matching)
			match := SearchMatch{
				File:       filename,
				LineNumber: lineNum,
				Column:     column,
				Line:       content,
				Match:      req.Query, // Simplified - could extract actual match
			}

			results.Matches = append(results.Matches, match)
			results.TotalMatches++
		}
	}

	results.TotalFiles = len(fileSet)
	for file := range fileSet {
		results.Files = append(results.Files, file)
	}

	return results, nil
}
