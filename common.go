package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sdwolfe32/ANIRip/anirip"
)

// Trims the first couple seconds off of the video to remove any logos
func trimMKV(adLength int, engineDir, tempDir string) error {
	// Removes a stale temp files to avoid conflcts in func
	os.Remove(tempDir + "\\" + "untrimmed.episode.mkv")
	os.Remove(tempDir + "\\" + "split.episode-001.mkv")
	os.Remove(tempDir + "\\" + "prefix.episode.mkv")
	os.Remove(tempDir + "\\" + "split.episode-002.mkv")
	os.Remove(tempDir + "\\" + "list.episode.txt")

	// Recursively retries rename to temp filename before execution
	if err := anirip.Rename(tempDir+"\\episode.mkv", tempDir+"\\untrimmed.episode.mkv", 10); err != nil {
		return err
	}

	// Finds the clis we need for trimming
	ffmpeg, err := filepath.Abs(engineDir + "\\ffmpeg.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find ffmpeg.exe in \\" + engineDir + "\\ directory", Err: err}
	}
	mkvmerge, err := filepath.Abs(engineDir + "\\mkvmerge.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find ffmpeg.exe in \\" + engineDir + "\\ directory", Err: err}
	}

	// Creates the command too split the meat of the video from the first ad chunk
	cmd := exec.Command(mkvmerge,
		"--split", "timecodes:"+anirip.MStoTimecode(adLength),
		"-o", "split.episode.mkv",
		"untrimmed.episode.mkv",
	)
	cmd.Dir = tempDir // Sets working directory to temp so our halves end up there

	// Executes the command
	_, err = cmd.Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while splitting the episode", Err: err}
	}

	// We need to ask ffprobe explicitly for an exact framerate because ffmpeg's auto
	// frame reader thinks 30.30 frames is 30.3, resulting in frame jumps at end of prefix
	frameRate, err := getVideoFrameRate("untrimmed.episode.mkv", engineDir, tempDir)
	if err != nil {
		return err
	}
	frameRateString := strconv.FormatFloat(frameRate, 'f', 8, 64)

	// Executes the fine intro trim and waits for the command to finish
	cmd = exec.Command(ffmpeg,
		"-ss", anirip.MStoTimecode(adLength), // Exact timestamp of the ad endings
		"-i", "split.episode-001.mkv",
		"-crf", "8",
		"-vsync", "1",
		"-r", frameRateString,
		"-c:a", "aac", "-y", // Use AAC as audio codec to match video.mkv
		"prefix.episode.mkv")
	cmd.Dir = tempDir // Sets working directory to temp

	// Executes the command
	_, err = cmd.Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while creating the prefix clip", Err: err}
	}

	// Creates a text file containing the file names of the 2 files created above
	fileListBytes := []byte("file 'prefix.episode.mkv'\r\nfile 'split.episode-002.mkv'")
	if err = ioutil.WriteFile(tempDir+"\\"+"list.episode.txt", fileListBytes, 0644); err != nil {
		return anirip.Error{Message: "There was an error while creating list.episode.txt", Err: err}
	}

	// Executes the merge of our two temporary files
	cmd = exec.Command(ffmpeg,
		"-f", "concat",
		"-i", "list.episode.txt",
		"-c", "copy", "-y",
		"episode.mkv")
	cmd.Dir = tempDir // Sets working directory to temp

	// Executes the command
	_, err = cmd.Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while merging video and prefix", Err: err}
	}

	// Removes the temporary files we created as they are no longer needed
	os.Remove(tempDir + "\\" + "untrimmed.episode.mkv")
	os.Remove(tempDir + "\\" + "split.episode-001.mkv")
	os.Remove(tempDir + "\\" + "prefix.episode.mkv")
	os.Remove(tempDir + "\\" + "split.episode-002.mkv")
	os.Remove(tempDir + "\\" + "list.episode.txt")
	return nil
}

// Merges a VIDEO.mkv and a VIDEO.ass
func mergeSubtitles(audioLang, subtitleLang, engineDir, tempDir string) error {
	// Removes a stale temp files to avoid conflcts in func
	os.Remove(tempDir + "\\unmerged.episode.mkv")

	// Recursively retries rename to temp filename before execution
	if err := anirip.Rename(tempDir+"\\episode.mkv", tempDir+"\\unmerged.episode.mkv", 10); err != nil {
		return err
	}

	path, err := filepath.Abs(engineDir + "\\ffmpeg.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find ffmpeg.exe in \\" + engineDir + "\\ directory", Err: err}
	}

	// Creates the command which we will use to merge our subtitles and video
	cmd := new(exec.Cmd)
	if subtitleLang == "" {
		cmd = exec.Command(path,
			"-i", "unmerged.episode.mkv",
			"-c:v", "copy",
			"-c:a", "copy",
			"-metadata:s:a:0", "language="+audioLang, // sets audio language to passed audioLang
			"-y", "episode.mkv")
	} else {
		cmd = exec.Command(path,
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
	cmd.Dir = tempDir // Sets working directory to temp

	// Executes the command
	_, err = cmd.Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while merging subtitles", Err: err}
	}

	// Removes old temp files
	os.Remove(tempDir + "\\subtitles.episode.ass")
	os.Remove(tempDir + "\\unmerged.episode.mkv")
	return nil
}

// Cleans up the mkv, optimizing it for playback
func cleanMKV(engineDir, tempDir string) error {
	// Removes a stale temp file to avoid conflcts in func
	os.Remove(tempDir + "\\dirty.episode.mkv")

	// Recursively retries rename to temp filename before execution
	if err := anirip.Rename(tempDir+"\\episode.mkv", tempDir+"\\"+"dirty.episode.mkv", 10); err != nil {
		return err
	}

	// Finds the path of mkclean.exe so we can perform system calls on it
	path, err := filepath.Abs(engineDir + "\\mkclean.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find mkclean.exe in \\" + engineDir + "\\ directory", Err: err}
	}

	// Creates the command which we will use to clean our mkv to "video.clean.mkv"
	cmd := exec.Command(path,
		"--optimize",
		"dirty.episode.mkv",
		"episode.mkv")
	cmd.Dir = tempDir // Sets working directory to temp

	// Executes the command
	_, err = cmd.Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while optimizing our mkv", Err: err}
	}

	// Deletes the old, un-needed dirty mkv file
	os.Remove(tempDir + "\\dirty.episode.mkv")
	return nil
}

// Uses ffprobe to find the length of a video and returns it in ms
func getVideoLength(fileName, engineDir, tempDir string) (int, error) {
	// Gets the ffprobe path which we will use to figure out the video length
	ffprobe, err := filepath.Abs(engineDir + "\\ffprobe.exe")
	if err != nil {
		return 0, anirip.Error{Message: "Unable to find ffprobe.exe in \\" + engineDir + "\\ directory", Err: err}
	}

	// Asks for the length of our video
	cmd := exec.Command(ffprobe,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		fileName)
	cmd.Dir = tempDir // Sets working directory to temp

	// Executes the command
	output, err := cmd.Output()
	if err != nil {
		return 0, anirip.Error{Message: "There was an error measuring " + fileName, Err: err}
	}

	// Grabs the output and parses it to a float64
	length, err := strconv.ParseFloat(strings.Replace(string(output), "\r\n", "", -1), 64)
	if err != nil {
		return 0, anirip.Error{Message: "There was an error parsing the length of " + fileName, Err: err}
	}
	return int(length * 1000), nil
}

// Uses ffprobe to get the exact framerate of the video and returns it as a float64
func getVideoFrameRate(fileName, engineDir, tempDir string) (float64, error) {
	// Gets the ffprobe path which we will use to figure out the video length
	ffprobe, err := filepath.Abs(engineDir + "\\ffprobe.exe")
	if err != nil {
		return 0, anirip.Error{Message: "Unable to find ffprobe.exe in \\" + engineDir + "\\ directory", Err: err}
	}

	// Asks for the length of our video
	cmd := exec.Command(ffprobe,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=avg_frame_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		fileName)
	cmd.Dir = tempDir // Sets working directory to temp

	// Executes the command
	output, err := cmd.Output()
	if err != nil {
		return 0, anirip.Error{Message: "There was an error measuring " + fileName, Err: err}
	}

	// Output will be represented as a fraction which needs to be solved
	frameRateArray := strings.Split(strings.Replace(string(output), "\r\n", "", -1), "/")
	numerator, err := strconv.ParseFloat(frameRateArray[0], 64)
	if err != nil {
		return 0, anirip.Error{Message: "There was an error parsing the numerator of our framerate for " + fileName, Err: err}
	}
	denominator, err := strconv.ParseFloat(frameRateArray[1], 64)
	if err != nil {
		return 0, anirip.Error{Message: "There was an error parsing the denominator of our framerate for " + fileName, Err: err}
	}
	return numerator / denominator, nil
}

// Gets user input from the user and unmarshalls it into the input
func getStandardUserInput(prefixText string, input *string) error {
	fmt.Printf(prefixText)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		*input = scanner.Text()
		break
	}
	if err := scanner.Err(); err != nil {
		return anirip.Error{Message: "There was an error getting standard user input", Err: err}
	}
	return nil
}

// Blocks execution and waits for the user to press enter
func pause() {
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
