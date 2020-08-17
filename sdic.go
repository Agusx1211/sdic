package main

import (
	"fmt"
	"os"

	"bufio"
	"strings"

	"github.com/spf13/cobra"
)

func loadChunks(dicDir string, separator string) ([][]string, error) {
	// Define chunks slice
	chunks := make([][]string, 1)

	// Open dictionary file
	file, err := os.Open(dicDir)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Read all lines
	// split into chunks
	scanner := bufio.NewScanner(file)
	chunks = append(chunks, make([]string, 1))
	chunks[0] = append(chunks[0], "")
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
		return nil, err
	}

	return chunks[0 : len(chunks)-1], nil
}

func sdicMain(dicDir string, separator string) error {
	// Get chunks
	chunks, err := loadChunks(dicDir, separator)
	if err != nil {
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

func genRule(dicDir string, separator string, output string, rd int) error {
	// Get chunks
	chunks, err := loadChunks(dicDir, separator)
	if err != nil {
		return err
	}

	// Create files
	rfile, err := os.Create(output + ".rule")
	if err != nil {
		return err
	}
	defer rfile.Close()

	dfile, err := os.Create(output + ".dict")
	if err != nil {
		return err
	}
	defer dfile.Close()

	// Write rule (last chunk)
	_, err = rfile.WriteString(":\n")
	if err != nil {
		return err
	}

	indexes := make([]int, rd)
	ruleschunks := chunks[len(chunks)-rd:]
	for {
		// Generate candidate
		var str strings.Builder
		for i, bchunk := range ruleschunks {
			str.WriteString(bchunk[indexes[i]])
		}

		if str.String() != "" {
			for _, char := range str.String() {
				_, err := rfile.WriteString("$" + string(char))
				if err != nil {
					return err
				}
			}
			_, err := rfile.WriteString("\n")
			if err != nil {
				return err
			}
		}

		// Move indexes
		carry := false
		last := len(ruleschunks) - 1
		indexes[last] = indexes[last] + 1
		for i := last; i >= 0; i-- {
			if carry {
				indexes[i] = indexes[i] + 1
				carry = false
			}
			if indexes[i] == len(ruleschunks[i]) {
				indexes[i] = 0
				carry = true
			}
		}
		if carry {
			break
		}
	}

	// Write dict without last chunk
	for i, val := range chunks[0 : len(chunks)-rd] {
		for _, entry := range val {
			if entry != "" {
				_, err := dfile.WriteString(entry + "\n")
				if err != nil {
					return err
				}
			}
		}
		if i != len(chunks)-(rd+1) {
			_, err := dfile.WriteString(separator + "\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func sizeOf(dicDir string, separator string) error {
	// Get chunks
	chunks, err := loadChunks(dicDir, separator)
	if err != nil {
		return err
	}

	total := 1
	for _, chunk := range chunks {
		total = total * len(chunk)
	}

	fmt.Println(total)
	return nil
}

func main() {
	rootCmd := &cobra.Command{
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

			return sdicMain(dictDir, separator)
		},
	}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "rules",
		Short: "Generate rules for hashcat",
		RunE: func(cmd *cobra.Command, args []string) error {
			dictDir, err := cmd.Flags().GetString("dict")
			if err != nil {
				return err
			}
			separator, err := cmd.Flags().GetString("separator")
			if err != nil {
				return err
			}
			output, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}
			rd, err := cmd.Flags().GetInt("depth")
			if err != nil {
				return err
			}
			return genRule(dictDir, separator, output, rd)
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "size",
		Short: "Compute set of possible candidates",
		RunE: func(cmd *cobra.Command, args []string) error {
			dictDir, err := cmd.Flags().GetString("dict")
			if err != nil {
				return err
			}
			separator, err := cmd.Flags().GetString("separator")
			if err != nil {
				return err
			}

			return sizeOf(dictDir, separator)
		},
	})
	rootCmd.PersistentFlags().StringP("dict", "d", "", "Dictionary file")
	rootCmd.PersistentFlags().StringP("separator", "s", "<---Chunk--->", "Chunk separator")
	rootCmd.PersistentFlags().StringP("output", "o", "./gen_rules", "Outputs for rule generator")
	rootCmd.PersistentFlags().IntP("depth", "", 2, "Depth for rule generator")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
