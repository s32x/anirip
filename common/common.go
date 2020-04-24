package common /* import "s32x.com/anirip/common" */

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var illegalChars = []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}

const pathSep = string(os.PathSeparator)

// Delete removes a file from the system
func Delete(a ...string) error { return os.Remove(strings.Join(a, pathSep)) }

// Rename renames the source to the desired destination file name and
// recursively retries i times if there are any issues
func Rename(src, dst string, i int) error {
	if err := os.Rename(src, dst); err != nil {
		if i > 0 {
			return Rename(src, dst, i-1)
		}
		return fmt.Errorf("renaming file: %w", err)
	}
	return nil
}

func Move(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}

	// Defer executes in bottom-top order, so the file will close before it is removed
	defer Delete(src)
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}

	defer dest.Close()
	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("copying source to destination file: %w", err)
	}
	return nil
}

// GenerateEpisodeFilename constructs an episode filename and returns the
// filename fully sanitized
func GenerateEpisodeFilename(show string, season int, episode float64, desc string) string {
	ep := fmt.Sprintf("%g", episode)
	if episode < 10 {
		ep = "0" + ep // Prefix a zero to episode
	}
	return CleanFilename(fmt.Sprintf("%s - S%sE%s - %s", show,
		fmt.Sprintf("%02d", season), ep, desc))
}

// CleanFilename cleans the filename of any illegal file characters to prevent
// write errors
func CleanFilename(name string) string {
	for _, bad := range illegalChars {
		name = strings.Replace(name, bad, "", -1)
	}
	return strings.Replace(name, "  ", " ", -1)
}
