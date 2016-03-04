package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sdwolfe32/ANIRip/anirip"
)

// Trims the first couple seconds off of the video to remove any logos
func trimMKV(fileName string, adLength, estKeyFrame int) error {
	// Recursively retries rename to temp filename before execution
	if err := os.Rename("temp\\"+fileName+".mkv", "temp\\untrimmed."+fileName+".mkv"); err != nil {
		trimMKV(fileName, adLength, estKeyFrame)
	}

	// Store the untrimmed video length so we can find the video prefix length later
	untrimmedLength, err := getVideoLength("temp\\untrimmed." + fileName + ".mkv")
	if err != nil {
		return err
	}

	// Finds ffmpeg so we can call system commands on it
	ffmpeg, err := exec.LookPath("engine\\ffmpeg.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find ffmpeg.exe in \\engine\\ directory", Err: err}
	}

	// Calculates the keyframe offsets for trimming the meat of the video
	keyFrameOffset := float64(estKeyFrame) / 1000
	keyFrameOffsetString := strconv.FormatFloat(keyFrameOffset, 'f', 3, 64)

	// Executes the rough frame trimming and waits for command to finish
	_, err = exec.Command(ffmpeg,
		"-ss", keyFrameOffsetString,
		"-i", "temp\\untrimmed."+fileName+".mkv",
		"-c:v", "copy",
		"-c:a", "copy", "-y",
		"temp\\video."+fileName+".mkv").Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while creating the video clip", Err: err}
	}

	// Gets the new video length and calculates the prefix length based on the sizes
	videoLength, err := getVideoLength("temp\\video." + fileName + ".mkv")
	if err != nil {
		return err
	}
	keyFrameGap := (untrimmedLength - videoLength) - adLength - 0040

	// Calculates the intro offsets we will use and represents it as a string
	trueOffset := float64(adLength) / 1000
	trueOffsetString := strconv.FormatFloat(trueOffset, 'f', 3, 64)
	gapOffset := float64(keyFrameGap) / 1000
	gapOffsetString := strconv.FormatFloat(gapOffset, 'f', 3, 64)

	// Executes the fine intro trim and waits for the command to finish
	_, err = exec.Command(ffmpeg,
		"-ss", trueOffsetString, // Exact timestamp of the ad endings
		"-i", "temp\\untrimmed."+fileName+".mkv",
		"-t", gapOffsetString, // The exact time between ad ending and frame next keyframe
		"-crf", "5",
		"-vsync", "1",
		"-r", "24",
		"-c:a", "aac", "-y", // Use AAC as audio codec to match video.mkv
		"temp\\prefix."+fileName+".mkv").Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while creating the prefix clip", Err: err}
	}

	// Creates a text file containing the file names of the 2 files created above
	fileListBytes := []byte("file 'temp\\prefix." + fileName + ".mkv'\r\nfile 'temp\\video." + fileName + ".mkv'")
	if err = ioutil.WriteFile("temp\\list."+fileName+".txt", fileListBytes, 0644); err != nil {
		return anirip.Error{Message: "There was an error while creating list." + fileName + ".txt", Err: err}
	}

	// Executes the merge of our two temporary files
	_, err = exec.Command(ffmpeg,
		"-f", "concat",
		"-i", "temp\\list."+fileName+".txt",
		"-c:v", "copy",
		"-c:a", "copy", "-y",
		"temp\\"+fileName+".mkv").Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while merging video and prefix", Err: err}
	}

	// Removes the temporary files we created as they are no longer neede
	os.Remove("temp\\untrimmed." + fileName + ".mkv")
	os.Remove("temp\\prefix." + fileName + ".mkv")
	os.Remove("temp\\video." + fileName + ".mkv")
	os.Remove("temp\\list." + fileName + ".txt")
	return nil
}

// Merges a VIDEO.mkv and a VIDEO.ass
func mergeSubtitles(fileName, audioLang, subtitleLang string) error {
	// Recursively retries rename to temp filename before execution
	if err := os.Rename("temp\\"+fileName+".mkv", "temp\\unmerged."+fileName+".mkv"); err != nil {
		mergeSubtitles(fileName, audioLang, subtitleLang)
	}

	path, err := exec.LookPath("engine\\ffmpeg.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find ffmpeg.exe in \\engine\\ directory", Err: err}
	}

	// Creates the command which we will use to merge our subtitles and video
	cmd := exec.Command(path,
		"-i", "temp\\unmerged."+fileName+".mkv",
		"-f", "ass",
		"-i", "temp\\"+fileName+".ass",
		"-c:v", "copy",
		"-c:a", "copy",
		"-metadata:s:a:0", "language="+audioLang, // sets audio language to passed audioLang
		"-metadata:s:s:0", "language="+subtitleLang, // sets subtitle language to subtitleLang
		"-disposition:s:0", "default",
		"-y", "temp\\"+fileName+".mkv")

	// Executes the merge command and waits for a response
	err = cmd.Start()
	if err != nil {
		return anirip.Error{Message: "There was an error while executing our merger", Err: err}
	}
	err = cmd.Wait()
	if err != nil {
		return anirip.Error{Message: "There was an error while merging", Err: err}
	}

	// Removes old temp files
	os.Remove("temp\\" + fileName + ".ass")
	os.Remove("temp\\unmerged." + fileName + ".mkv")
	return nil
}

// Cleans up the mkv, optimizing it for playback
func cleanMKV(fileName string) error {
	// Recursively retries rename to temp filename before execution
	if err := os.Rename("temp\\"+fileName+".mkv", "temp\\dirty."+fileName+".mkv"); err != nil {
		cleanMKV(fileName)
	}

	// Finds the path of mkclean.exe so we can perform system calls on it
	path, err := exec.LookPath("engine\\mkclean.exe")
	if err != nil {
		return anirip.Error{Message: "Unable to find mkclean.exe in \\engine\\ directory", Err: err}
	}

	// Creates the command which we will use to clean our mkv to "video.clean.mkv"
	_, err = exec.Command(path,
		"--optimize",
		"temp\\dirty."+fileName+".mkv",
		"temp\\"+fileName+".mkv").Output()
	if err != nil {
		return anirip.Error{Message: "There was an error while optimizing our mkv", Err: err}
	}

	// Deletes the old, un-needed dirty mkv file
	os.Remove("temp\\dirty." + fileName + ".mkv")
	return nil
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

// Uses ffprobe to find the length of a video and returns it in ms
func getVideoLength(videoPath string) (int, error) {
	// Gets the ffprobe path which we will use to figure out the video length
	ffprobe, err := exec.LookPath("engine\\ffprobe.exe")
	if err != nil {
		return 0, anirip.Error{Message: "Unable to find ffprobe.exe in \\engine\\ directory", Err: err}
	}

	// Asks for the length of our video
	output, err := exec.Command(ffprobe,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath).Output()
	if err != nil {
		return 0, anirip.Error{Message: "There was an error measuring " + videoPath, Err: err}
	}

	// Grabs the output and parses it to a float64
	length, err := strconv.ParseFloat(strings.Replace(string(output), "\r\n", "", -1), 64)
	if err != nil {
		return 0, anirip.Error{Message: "There was an error parsing the length of " + videoPath, Err: err}
	}
	return int(length * 1000), nil
}

// Blocks execution and waits for the user to press enter
func pause() {
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
