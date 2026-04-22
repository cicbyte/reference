package repo

import (
	"bufio"
	"os"
	"strings"
)

const refIgnoreEntry = ".reference/"

func EnsureGitignore(projectDir string) error {
	gitignorePath := projectDir + "/.gitignore"
	entries, err := readLines(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(gitignorePath, []byte(refIgnoreEntry+"\n"), 0644)
		}
		return err
	}

	for _, line := range entries {
		if strings.TrimSpace(line) == ".reference/" || strings.TrimSpace(line) == ".reference" {
			return nil
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if len(entries) > 0 && !strings.HasSuffix(entries[len(entries)-1], "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			f.Close()
			return err
		}
	}
	_, err = f.WriteString(refIgnoreEntry + "\n")
	f.Close()
	if err != nil {
		return err
	}

	// Windows 追加写入可能产生 CRLF，重新读取并写入以确保 LF
	content, readErr := os.ReadFile(gitignorePath)
	if readErr != nil {
		return nil
	}
	cleaned := strings.ReplaceAll(string(content), "\r\n", "\n")
	return os.WriteFile(gitignorePath, []byte(cleaned), 0644)
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
