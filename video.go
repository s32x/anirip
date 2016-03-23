package main

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sdwolfe32/ANIRip/anirip"
)

// Trims the first couple seconds off of the video to remove any logos
func trimMKV(adLength int) error {
	// Removes a stale temp files to avoid conflcts in func
	os.Remove("untrimmed.episode.mkv")
	os.Remove("split.episode-001.mkv")
	os.Remove("prefix.episode.mkv")
	os.Remove("split.episode-002.mkv")
	os.Remove("list.episode.txt")

	// Recursively retries rename to temp filename before execution
	if err := anirip.Rename("episode.mkv", "untrimmed.episode.mkv", 10); err != nil {
		return err
	}

	// Executes the command too split the meat of the video from the first ad chunk
	if err := exec.Command("mkvmerge",
		"--split", "timecodes:"+anirip.MStoTimecode(adLength),
		"-o", "split.episode.mkv",
		"untrimmed.episode.mkv").Run(); err != nil {
		return anirip.Error{Message: "There was an error while splitting the episode", Err: err}
	}

	// Executes the fine intro trim and waits for the command to finish
	if err := exec.Command("ffmpeg",
		"-i", "split.episode-001.mkv",
		"-ss", anirip.MStoTimecode(adLength), // Exact timestamp of the ad endings
		"-c:v", "h264",
		"-crf", "15",
		"-preset", "slow",
		"-c:a", "copy", "-y", // Use AAC as audio codec to match video.mkv
		"prefix.episode.mkv").Run(); err != nil {
		return anirip.Error{Message: "There was an error while creating the prefix clip", Err: err}
	}

	// Creates a text file containing the file names of the 2 files created above
	fileListBytes := []byte("file 'prefix.episode.mkv'\r\nfile 'split.episode-002.mkv'")
	if err := ioutil.WriteFile("list.episode.txt", fileListBytes, 0644); err != nil {
		return anirip.Error{Message: "There was an error while creating list.episode.txt", Err: err}
	}

	// Executes the merge of our two temporary files
	if err := exec.Command("ffmpeg",
		"-f", "concat",
		"-i", "list.episode.txt",
		"-c", "copy", "-y",
		"episode.mkv").Run(); err != nil {
		return anirip.Error{Message: "There was an error while merging video and prefix", Err: err}
	}

	// Removes the temporary files we created as they are no longer needed
	os.Remove("untrimmed.episode.mkv")
	os.Remove("split.episode-001.mkv")
	os.Remove("prefix.episode.mkv")
	os.Remove("split.episode-002.mkv")
	os.Remove("list.episode.txt")
	return nil
}

// Merges a VIDEO.mkv and a VIDEO.ass
func mergeSubtitles(audioLang, subtitleLang string) error {
	// Removes a stale temp files to avoid conflcts in func
	os.Remove("unmerged.episode.mkv")

	// Recursively retries rename to temp filename before execution
	if err := anirip.Rename("episode.mkv", "unmerged.episode.mkv", 10); err != nil {
		return err
	}

	// Creates the command which we will use to merge our subtitles and video
	cmd := new(exec.Cmd)
	if subtitleLang == "" {
		cmd = exec.Command("ffmpeg",
			"-i", "unmerged.episode.mkv",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang, // sets audio language to passed audioLang
			"-y", "episode.mkv")
	} else {
		cmd = exec.Command("ffmpeg",
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

	// Executes the command
	if err := cmd.Run(); err != nil {
		return anirip.Error{Message: "There was an error while merging subtitles", Err: err}
	}

	// Removes old temp files
	os.Remove("subtitles.episode.ass")
	os.Remove("unmerged.episode.mkv")
	return nil
}

// Cleans up the mkv, optimizing it for playback
func cleanMKV() error {
	// Removes a stale temp file to avoid conflcts in func
	os.Remove("dirty.episode.mkv")

	// Recursively retries rename to temp filename before execution
	if err := anirip.Rename("episode.mkv", "dirty.episode.mkv", 10); err != nil {
		return err
	}

	// Executes the command which we will use to clean our mkv to "video.clean.mkv"
	if err := exec.Command("mkclean",
		"--optimize",
		"dirty.episode.mkv",
		"episode.mkv").Run(); err != nil {
		return anirip.Error{Message: "There was an error while optimizing our mkv", Err: err}
	}

	// Deletes the old, un-needed dirty mkv file
	os.Remove("dirty.episode.mkv")
	return nil
}
