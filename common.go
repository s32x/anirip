package main

import (
	"bufio"
	"fmt"
	"os"
)

// Gets user input from the user and unmarshalls it into the input
func GetStandardUserInput(prefixText string, input *string) error {
	fmt.Printf(prefixText)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		*input = scanner.Text()
		break
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
