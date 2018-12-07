package common /* import "s32x.com/anirip/common" */

import (
	"fmt"
	"os"
	"strings"
)

var illegalChars = []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}

// Rename renames the source to the desired destination
// file name and recursively retries i times if there
// are any issues
func Rename(src, dst string, i int) error {
	if err := os.Rename(src, dst); err != nil {
		if i > 0 {
			return Rename(src, dst, i-1)
		}
		return err
	}
	return nil
}

// GenerateEpisodeFilename constructs an episode filename
// and returns the filename fully sanitized
func GenerateEpisodeFilename(show string, season int, episode float64, desc string) string {
	ep := fmt.Sprintf("%g", episode)
	if episode < 10 {
		ep = "0" + ep // Prefix a zero to episode
	}
	return CleanFilename(fmt.Sprintf("%s - S%sE%s - %s", show, fmt.Sprintf("%02d", season), ep, desc))
}

// CleanFilename cleans the filename of any illegal file
// characters to prevent write errors
func CleanFilename(name string) string {
	for _, bad := range illegalChars {
		name = strings.Replace(name, bad, "", -1)
	}
	return strings.Replace(name, "  ", " ", -1)
}
