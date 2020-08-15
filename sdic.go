package main

import (
	"fmt"
	"os"

	"bufio"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sdic",
	Short: "Password candidate generator that combinates words from a dictionary divided in chunks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dictDir, err := cmd.Flags().GetString("dict")
		if err != nil {
			return err
		}
		separator, err := cmd.Flags().GetString("separator")
		if err != nil {
			return err
		}

		if separator == "" {
			separator = "<---Chunk--->"
		}

		return sdicMain(dictDir, separator)
	},
}

func init() {
	rootCmd.Flags().StringP("dict", "d", "", "Dictionary file")
	rootCmd.Flags().StringP("separator", "s", "", "Chunk separator")
}

func sdicMain(dicDir string, separator string) error {
	// Define chunks slice
	chunks := make([][]string, 1)

	// Open dictionary file
	file, err := os.Open(dicDir)
	if err != nil {
		return err
	}

	defer file.Close()

	// Read all lines
	// split into chunks
	scanner := bufio.NewScanner(file)
	chunks = append(chunks, make([]string, 1))
	cindex := 0
	for scanner.Scan() {
		text := scanner.Text()

		if text == separator {
			cindex = cindex + 1
			chunks = append(chunks, make([]string, 1))
		} else {
			// Search for duplicates
			exists := false
			for _, entry := range chunks[cindex] {
				if entry == text {
					exists = true
					break
				}
			}

			if !exists {
				chunks[cindex] = append(chunks[cindex], text)
			}
		}
	}

	// Close file
	if err := scanner.Err(); err != nil {
		return err
	}

	// Print all combinations in order
	indexes := make([]int, len(chunks))
	for {
		// Generate candidate
		var str strings.Builder
		for i := range chunks {
			str.WriteString(chunks[i][indexes[i]])
		}
		fmt.Println(str.String())

		// Move indexes
		carry := false
		last := len(chunks) - 1
		indexes[last] = indexes[last] + 1
		for i := last; i >= 0; i-- {
			if carry {
				indexes[i] = indexes[i] + 1
				carry = false
			}
			if indexes[i] == len(chunks[i]) {
				indexes[i] = 0
				carry = true
			}
		}
		if carry {
			break
		}
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
