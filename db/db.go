// Package db implements the database operations for the application
package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"webapp/models"

	"github.com/gofrs/flock"
)

// TODO: ensure all incoming data is trusted

// baseDir is the base directory to save the data
const baseDir = "db/data"

// fileLocks is a map of file locks to handle concurrent access
var fileLocks sync.Map

// getFileLock returns the file lock for the given file path
func getFileLock(filePath string) *flock.Flock {
	value, _ := fileLocks.LoadOrStore(filePath, flock.New(filePath))
	return value.(*flock.Flock)
}

// fileNameForNote generates a file name for the given note
func fileNameForNote(newNote *models.Note) string {
	return newNote.Text
}

// SaveDeposit saves a deposit to the database
func SaveDeposit(data *models.DepositData) error {
	if data.Note == nil {
		return fmt.Errorf("note is nil")
	}
	return saveNoteToFile(data.Note)
}

// SaveWithdrawal saves a withdrawal to the database
func SaveWithdrawal(data *models.WithdrawData) error {
	if data.NewNote == nil {
		return fmt.Errorf("new note is nil")
	}
	return saveNoteToFile(data.NewNote)
}

// saveNoteToFile saves a note to a file
func saveNoteToFile(data *models.Note) error {
	fileName := fileNameForNote(data)
	dir := filepath.Join(baseDir, fileName[len(fileName)-2:])
	filePath := filepath.Join(dir, fileName+".json")

	// Lock the directory before creating it
	dirLock := getFileLock(dir)
	err := dirLock.Lock()
	if err != nil {
		return fmt.Errorf("error locking directory: %w", err)
	}
	defer dirLock.Unlock()
	// Create the directory if it doesn't exist
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	// Now lock the file before writing
	fileLock := getFileLock(filePath)
	err = fileLock.Lock()
	if err != nil {
		return fmt.Errorf("error locking file: %w", err)
	}
	defer fileLock.Unlock()

	// Create an empty file if it doesn't exist
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		file.Close()
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}

// ExistNote checks if a note file exists in the database.
func ExistNote(n *models.Note) (bool, error) {
	fileName := fileNameForNote(n)
	dir := filepath.Join(baseDir, fileName[len(fileName)-2:])
	filePath := filepath.Join(dir, fileName+".json")

	// Lock the file before reading
	lock := getFileLock(filePath)
	err := lock.RLock()
	if err != nil {
		return false, fmt.Errorf("error locking file: %w", err)
	}
	defer lock.Unlock()

	// Check if the file exists
	_, err = os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File does not exist
		}
		return false, fmt.Errorf("error reading file: %w", err)
	}

	return true, nil
}

// DeleteNote deletes a note from the database
func DeleteNote(n *models.Note) error {
	fileName := fileNameForNote(n)
	dir := filepath.Join(baseDir, fileName[len(fileName)-2:])
	filePath := filepath.Join(dir, fileName+".json")

	// Lock the file before deleting
	lock := getFileLock(filePath)
	err := lock.Lock()
	if err != nil {
		return fmt.Errorf("error locking file: %w", err)
	}
	defer lock.Unlock()

	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}
	return nil
}
