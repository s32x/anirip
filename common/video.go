package common /* import "s32x.com/anirip/common" */

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type VideoProcessor struct{ tempDir string }

// NewVideoProcessor generates a new VideoProcessor that
// contains the location of our temporary directory
func NewVideoProcessor(tempDir string) *VideoProcessor {
	return &VideoProcessor{tempDir: tempDir}
}

// DumpHLS dumps an HLS Stream to the temporary directory
func (p *VideoProcessor) DumpHLS(url string) error {
	// Delete a previous incomplete episode file
	Delete(p.tempDir, "incomplete.episode.mkv")

	// Generate and execute the ffmpeg dump command
	cmd := exec.Command(
		findAbsoluteBinary("ffmpeg"),
		"-i", url,
		"-c", "copy", "incomplete.episode.mkv")
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running download command: %w", err)
	}

	// Rename the file since it's no longer incomplete
	// and return
	if err := Rename(p.tempDir+pathSep+"incomplete.episode.mkv",p.tempDir+pathSep+"episode.mkv", 10); err != nil {
		return fmt.Errorf("renaming incomplete episode: %w", err)
	}
	return nil
}

// MergeSubtitles merges the VIDEO.mkv and the VIDEO.ass
func (p *VideoProcessor) MergeSubtitles(audioLang, subtitleLang string) error {
	Delete(p.tempDir, "unmerged.episode.mkv")
	if err := Rename(p.tempDir+pathSep+"episode.mkv", p.tempDir+pathSep+"unmerged.episode.mkv", 10); err != nil {
		return fmt.Errorf("renaming unmerged episode: %w", err)
	}
	cmd := new(exec.Cmd)
	if subtitleLang == "" {
		cmd = exec.Command(
			findAbsoluteBinary("ffmpeg"),
			"-i", "unmerged.episode.mkv",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang,
			"-y", "episode.mkv")
	} else {
		cmd = exec.Command(
			findAbsoluteBinary("ffmpeg"),
			"-i", "unmerged.episode.mkv",
			"-f", "ass",
			"-i", "subtitles.episode.ass",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang,
			"-metadata:s:s:0", "language="+subtitleLang,
			"-disposition:s:0", "default",
			"-y", "episode.mkv")
	}
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running download command: %w", err)
	}
	Delete(p.tempDir, "subtitles.episode.ass")
	Delete(p.tempDir, "unmerged.episode.mkv")
	return nil
}

// findAbsoluteBinary attempts to search, find, and return the absolute path of
// the desired binary
func findAbsoluteBinary(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		path = name
	}
	path, err = filepath.Abs(path)
	if err != nil {
		path = name
	}
	return path
}
