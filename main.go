package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/pflag"
)

type CommandEntry struct {
	Command     string
	Count       int
	LastUsed    int64  // Unix timestamp in nanoseconds
	LastUsedStr string // Formatted date string
}

func main() {
	// Define command-line flags
	var includeDeleted bool
	var reverseOrder bool
	var printNull bool
	var cwdDir string
	var session string
	var dbPath string
	var fieldSeparator string
	var ansiEnabled bool
	var header bool
	var headerLast bool

	pflag.BoolVarP(&includeDeleted, "include-deleted", "d", false, "Include deleted commands")
	pflag.BoolVarP(&reverseOrder, "reverse", "r", false, "Reverse the sort order (oldest first)")
	pflag.BoolVarP(&printNull, "print0", "0", false, "Use null character as record separator")
	pflag.StringVarP(&cwdDir, "cwd", "c", "", "limit search to a specific directory")
	pflag.StringVarP(&session, "session", "s", "", "limit search to a specific session")
	pflag.StringVarP(&dbPath, "db", "", "", "Path to the database file")
	pflag.StringVarP(&fieldSeparator, "fieldsep", "f", "║", "Field separator for output")
	pflag.BoolVarP(&ansiEnabled, "ansi", "a", false, "Enable ANSI colors")
	pflag.BoolVar(&header, "header", false, "Print header before results")
	pflag.BoolVar(&headerLast, "header-last", false, "Print header after results")

	pflag.Lookup("cwd").NoOptDefVal = getCurrentWorkingDir()
	pflag.Lookup("session").NoOptDefVal = os.Getenv("ATUIN_SESSION")

	pflag.Parse()

	// Get the database path
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		dbPath = filepath.Join(homeDir, ".local", "share", "atuin", "history.db")
	}

	// Process the database and output results
	if err := processHistory(dbPath, includeDeleted, reverseOrder, printNull, cwdDir, session, fieldSeparator, ansiEnabled, header, headerLast); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "dbPath: %v\n", dbPath)
		fmt.Fprintf(os.Stderr, "cwdDir: %v\n", cwdDir)
		fmt.Fprintf(os.Stderr, "session: %v\n", session)
		fmt.Fprintf(os.Stderr, "includeDeleted: %v\n", includeDeleted)
		fmt.Fprintf(os.Stderr, "reverseOrder: %v\n", reverseOrder)
		fmt.Fprintf(os.Stderr, "printNull: %v\n", printNull)
		fmt.Fprintf(os.Stderr, "header: %v\n", header)
		fmt.Fprintf(os.Stderr, "ansiEnabled: %v\n", ansiEnabled)
		fmt.Fprintf(os.Stderr, "header-last: %v\n", headerLast)
		os.Exit(1)
	}
}

func getCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		os.Exit(1)
	}
	return dir
}

func processHistory(dbPath string, includeDeleted, reverseOrder, printNull bool, cwdDir string, session string, fieldSeparator string, ansiEnabled bool, header bool, headerLast bool) error {
	whereClausePresent := false

	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Build the query
	query := "SELECT command, timestamp, deleted_at FROM history"
	var args []interface{}
	if !includeDeleted {
		query += " WHERE deleted_at IS NULL"
		whereClausePresent = true
	}
	if cwdDir != "" {
		if !whereClausePresent {
			query += " WHERE"
		} else {
			query += " AND"
		}
		query += " cwd = ?"
		args = append(args, cwdDir)
	}
	if session != "" {
		if !whereClausePresent {
			query += " WHERE"
		} else {
			query += " AND"
		}
		query += " session = ?"
		args = append(args, session)
	}

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	// Process the results
	commandMap := make(map[string]*CommandEntry)
	maxCount := 0 // Track the maximum count for padding

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

		// Update max count if needed
		if entry.Count > maxCount {
			maxCount = entry.Count
		}

		// Update the last used timestamp if this is more recent
		if timestamp > entry.LastUsed {
			entry.LastUsed = timestamp
			// Format the timestamp as YYYY-MM-DD hh:mm:ss
			// Timestamp is in nanoseconds since epoch
			t := time.Unix(0, timestamp)
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

	// Calculate the width needed for the count column
	countWidth := len(fmt.Sprintf("%d", maxCount))
	// Ensure header width for 'COUNT'
	if countWidth < len("COUNT") && (header || headerLast) {
		countWidth = len("COUNT")
	}
	// Determine column width for time strings (e.g. "2025-04-21 12:34:56")
	timeWidth := len("2006-01-02 15:04:05")

	// ANSI color codes (conditionally enabled)
	purpleColor, blueColor, greenColor, resetColor := "", "", "", ""
	if ansiEnabled {
		purpleColor = "\033[95m" // bright purple
		blueColor = "\033[94m"   // bright blue
		greenColor = "\033[92m"  // bright green
		resetColor = "\033[0m"
	}

	// Print header if requested
	if header {
		// Header: TIME, padded COUNT, COMMAND with colors (left-justified)
		fmt.Printf("%s%-*s%s %s%s%s %s%-*s%s %s%s%s %s%c",
			purpleColor, timeWidth, "TIME", resetColor,
			blueColor, fieldSeparator, resetColor,
			greenColor, countWidth, "COUNT", resetColor,
			blueColor, fieldSeparator, resetColor,
			"COMMAND", '\n')
	}

	// Output the results
	for _, entry := range commands {
		recordSeparator := '\n'
		if printNull {
			recordSeparator = 0
		}

		// Colored output: purple timestamp, blue separators, green count
		fmt.Printf("%s%s%s %s%s%s %s%*d%s %s%s%s %s%c",
			purpleColor, entry.LastUsedStr, resetColor,
			blueColor, fieldSeparator, resetColor,
			greenColor, countWidth, entry.Count, resetColor,
			blueColor, fieldSeparator, resetColor,
			entry.Command,
			recordSeparator)
	}

	// Print header at end if requested
	if headerLast {
		// Header: TIME, padded COUNT, COMMAND with colors (left-justified)
		fmt.Printf("%s%-*s%s %s%s%s %s%-*s%s %s%s%s %s%c",
			purpleColor, timeWidth, "TIME", resetColor,
			blueColor, fieldSeparator, resetColor,
			greenColor, countWidth, "COUNT", resetColor,
			blueColor, fieldSeparator, resetColor,
			"COMMAND", '\n')
	}

	return nil
}
