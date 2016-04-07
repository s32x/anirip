package anirip

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Rename func that retries 10 times before returning an error
func Rename(sourcesFile, destinationFile string, i int) error {
	// Attempts a rename and if it fails, it will retry i times
	if err := os.Rename(sourcesFile, destinationFile); err != nil {
		if i > 0 {
			Rename(sourcesFile, destinationFile, i-1)
		}
		return Error{Message: "There was an error renaming " + sourcesFile + " to " + destinationFile, Err: err}
	}
	return nil
}

// Constructs an episode file name and returns the file name cleaned
func GenerateEpisodeFileName(showTitle string, seasonNumber int, episodeNumber float64, description string) string {
	// Pads season number with a 0 if it's less than 10
	seasonNumberString := strconv.Itoa(seasonNumber)
	if seasonNumber < 10 {
		seasonNumberString = "0" + strconv.Itoa(seasonNumber)
	}

	// Pads episode number with a 0 if it's less than 10
	episodeNumberString := strconv.FormatFloat(episodeNumber, 'f', -1, 64)
	if episodeNumber < 10 {
		episodeNumberString = "0" + strconv.FormatFloat(episodeNumber, 'f', -1, 64)
	}

	// Constructs episode file name and returns it
	fileName := showTitle + " - S" + seasonNumberString + "E" + episodeNumberString + " - " + description
	return CleanFileName(fileName)
}

// Cleans the new file/folder name so there won't be any write issues
func CleanFileName(fileName string) string {
	newFileName := fileName // Strips out any illegal characters and returns our new file name
	for _, illegalChar := range []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"} {
		newFileName = strings.Replace(newFileName, illegalChar, "", -1)
	}
	newFileName = strings.Replace(newFileName, "  ", " ", -1) // Replaces double spaces with a single space
	return newFileName
}

// Attempts to search, find, and return the absolute path of a given binary
func FindAbsoluteBinary(binaryName string) string {
	// Searches for the binary whether it be in the path or relative
	lookPath, err := exec.LookPath(binaryName)
	if err != nil {
		lookPath = binaryName
	}

	// Makes our path absolute regardless of where it is
	absPath, err := filepath.Abs(lookPath)
	if err != nil {
		absPath = binaryName
	}

	// Returns the absolute path to the binary we're looking to utilize
	return absPath
}
