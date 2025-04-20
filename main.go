package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CommandEntry struct {
	Command     string
	Count       int
	LastUsed    int64 // Unix timestamp
	LastUsedStr string // Formatted date string
}

func main() {
	// Define command-line flags
	includeDeleted := flag.Bool("include-deleted", false, "Include deleted commands")
	reverseOrder := flag.Bool("reverse", false, "Reverse the sort order (oldest first)")
	flag.Parse()

	// Get the database path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(homeDir, ".local", "share", "atuin", "history.db")

	// Process the database and output results
	if err := processHistory(dbPath, *includeDeleted, *reverseOrder); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processHistory(dbPath string, includeDeleted, reverseOrder bool) error {
	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Build the query
	query := "SELECT command, timestamp, deleted_at FROM history"
	if !includeDeleted {
		query += " WHERE deleted_at IS NULL"
	}

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	// Process the results
	commandMap := make(map[string]*CommandEntry)
	for rows.Next() {
		var command string
		var timestamp int64
		var deletedAt sql.NullInt64

		if err := rows.Scan(&command, &timestamp, &deletedAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Skip if deleted and we're not including deleted commands
		if !includeDeleted && deletedAt.Valid {
			continue
		}

		// Update the command entry
		entry, exists := commandMap[command]
		if !exists {
			entry = &CommandEntry{
				Command:  command,
				Count:    0,
				LastUsed: 0,
			}
			commandMap[command] = entry
		}

		// Increment the count
		entry.Count++

		// Update the last used timestamp if this is more recent
		if timestamp > entry.LastUsed {
			entry.LastUsed = timestamp
			// Format the timestamp as YYYY-MM-DD hh:mm:ss
			t := time.Unix(timestamp/1000000, 0) // Convert microseconds to seconds
			entry.LastUsedStr = t.Format("2006-01-02 15:04:05")
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	// Convert map to slice for sorting
	var commands []*CommandEntry
	for _, entry := range commandMap {
		commands = append(commands, entry)
	}

	// Sort by last used timestamp (most recent first by default)
	sort.Slice(commands, func(i, j int) bool {
		if reverseOrder {
			return commands[i].LastUsed < commands[j].LastUsed
		}
		return commands[i].LastUsed > commands[j].LastUsed
	})

	// Output the results
	for _, entry := range commands {
		fmt.Printf("%s │ %d │ %s%c", entry.LastUsedStr, entry.Count, entry.Command, 0)
	}

	return nil
}
