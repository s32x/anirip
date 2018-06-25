package common

import (
	"os"
	"os/exec"
	"path/filepath"
)

const pathSep = string(os.PathSeparator)

type VideoProcessor struct{ tempDir string }

// NewVideoProcessor generates a new VideoProcessor that
// contains the location of our temporary directory
func NewVideoProcessor(tempDir string) *VideoProcessor {
	return &VideoProcessor{tempDir: tempDir}
}

// DumpHLS dumps an HLS Stream to the temporary directory
func (p *VideoProcessor) DumpHLS(url string) error {
	// Remove a previous incomplete episode file
	os.Remove(p.tempDir + pathSep + "incomplete.episode.mkv")

	// Generate and execute the ffmpeg dump command
	cmd := exec.Command(
		findAbsoluteBinary("ffmpeg"),
		"-i", url,
		"-c", "copy",
		"incomplete.episode.mkv")
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Rename the file since it's no longer incomplete
	// and return
	return Rename(p.tempDir+pathSep+"incomplete.episode.mkv", p.tempDir+pathSep+"episode.mkv", 10)
}

// CleanMKV cleans the MKV metadata using mkclean
func (p *VideoProcessor) CleanMKV() error {
	// Rename the episode file before cleaning
	if err := Rename(p.tempDir+pathSep+"episode.mkv", p.tempDir+pathSep+"dirty.episode.mkv", 10); err != nil {
		return err
	}

	// Generate and execute the mkclean clean command
	// which will create a new copy of the file
	cmd := exec.Command(
		findAbsoluteBinary("mkclean"),
		"dirty.episode.mkv", "episode.mkv")
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Remove the old dirty (uncleaned) file and return
	return os.Remove(p.tempDir + pathSep + "dirty.episode.mkv")
}

// MergeSubtitles merges the VIDEO.mkv and the VIDEO.ass
func (p *VideoProcessor) MergeSubtitles(audioLang, subtitleLang string) error {
	os.Remove(p.tempDir + pathSep + "unmerged.episode.mkv")
	if err := Rename(p.tempDir+pathSep+"episode.mkv", p.tempDir+pathSep+"unmerged.episode.mkv", 10); err != nil {
		return err
	}
	cmd := new(exec.Cmd)
	if subtitleLang == "" {
		cmd = exec.Command(findAbsoluteBinary("ffmpeg"),
			"-i", "unmerged.episode.mkv",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang, // sets audio language to passed audioLang
			"-y", "episode.mkv")
	} else {
		cmd = exec.Command(findAbsoluteBinary("ffmpeg"),
			"-i", "unmerged.episode.mkv",
			"-f", "ass",
			"-i", "subtitles.episode.ass",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang, // sets audio language to passed audioLang
			"-metadata:s:s:0", "language="+subtitleLang, // sets subtitle language to subtitleLang
			"-disposition:s:0", "default",
			"-y", "episode.mkv")
	}
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}
	os.Remove(p.tempDir + pathSep + "subtitles.episode.ass")
	os.Remove(p.tempDir + pathSep + "unmerged.episode.mkv")
	return nil
}

// findAbsoluteBinary attempts to search, find, and
// return the absolute path of the desired binary
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
