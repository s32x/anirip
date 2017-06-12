package anirip

import (
	"os"
	"os/exec"
)

const pathSep = string(os.PathSeparator)

// VideoProcessor contains generic functionality needed for processing video
type VideoProcessor interface {
	DumpHLS(url string) error
	CleanMKV() error
	MergeSubtitles(audioLang, subLang string) error
}

type videoProcessor struct{ tempDir string }

// NewVideoProcessor generates a new VideoProcessor that contains the location
// of our temporary directory
func NewVideoProcessor(tempDir string) VideoProcessor {
	return &videoProcessor{tempDir: tempDir}
}

// DumpStream dumps an HLS stream to the temporary directory
func (p *videoProcessor) DumpHLS(url string) error {
	os.Remove(p.tempDir + pathSep + "incomplete.episode.mkv")
	cmd := exec.Command(FindAbsoluteBinary("ffmpeg"), "-i", url, "-c", "copy", "incomplete.episode.mkv")
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return Rename(p.tempDir+pathSep+"incomplete.episode.mkv", p.tempDir+pathSep+"episode.mkv", 10)
}

// cleanMKV cleans the MKV metadata using mkclean
func (p *videoProcessor) CleanMKV() error {
	if err := Rename(p.tempDir+pathSep+"episode.mkv", p.tempDir+pathSep+"dirty.episode.mkv", 10); err != nil {
		return err
	}
	cmd := exec.Command(FindAbsoluteBinary("mkclean"), "dirty.episode.mkv", "episode.mkv")
	cmd.Dir = p.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}
	os.Remove(p.tempDir + pathSep + "dirty.episode.mkv")
	return nil
}

// MergeSubtitles merges the VIDEO.mkv and the VIDEO.ass
func (p *videoProcessor) MergeSubtitles(audioLang, subtitleLang string) error {
	os.Remove(p.tempDir + pathSep + "unmerged.episode.mkv")
	if err := Rename(p.tempDir+pathSep+"episode.mkv", p.tempDir+pathSep+"unmerged.episode.mkv", 10); err != nil {
		return err
	}
	cmd := new(exec.Cmd)
	if subtitleLang == "" {
		cmd = exec.Command(FindAbsoluteBinary("ffmpeg"),
			"-i", "unmerged.episode.mkv",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang, // sets audio language to passed audioLang
			"-y", "episode.mkv")
	} else {
		cmd = exec.Command(FindAbsoluteBinary("ffmpeg"),
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
