package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/boyter/scc/v3/processor"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var sccOnce sync.Once

type SCCLanguageStat struct {
	Type       string `json:"type,omitempty"`
	Language   string `json:"languages"`
	Files      int64  `json:"files"`
	Lines      int64  `json:"lines"`
	Code       int64  `json:"code"`
	Comments   int64  `json:"comments"`
	Blanks     int64  `json:"blanks"`
	Complexity int64  `json:"complexity"`
}

type SCCFileStat struct {
	Type       string `json:"type,omitempty"`
	Filename   string `json:"filename"`
	Language   string `json:"language"`
	Location   string `json:"location"`
	Code       int64  `json:"code"`
	Complexity int64  `json:"complexity"`
}

func initSCC() {
	sccOnce.Do(func() {
		processor.ProcessConstants()
	})
}

var sccSkipDirs = map[string]bool{".git": true, "vendor": true, "node_modules": true, ".svn": true, ".hg": true}

// RunSCC 对指定路径运行代码统计，返回按语言的汇总和文件级统计
func RunSCC(repoPath string) ([]SCCLanguageStat, []SCCFileStat, error) {
	initSCC()

	cleanPath := filepath.Clean(repoPath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("路径不存在: %w", err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("不是目录: %s", cleanPath)
	}

	filePaths, err := collectFiles(cleanPath)
	if err != nil {
		return nil, nil, err
	}

	results := processFiles(cleanPath, filePaths)
	langStats, fileStats := aggregateResults(results)
	return langStats, fileStats, nil
}

func collectFiles(root string) ([]string, error) {
	var filePaths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if sccSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		filePaths = append(filePaths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %w", err)
	}
	return filePaths, nil
}

type fileResult struct {
	language   string
	filename   string
	location   string
	lines      int64
	code       int64
	comment    int64
	blank      int64
	complexity int64
}

func processFiles(cleanPath string, filePaths []string) []fileResult {
	results := make([]fileResult, 0, len(filePaths))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 32)

	for _, fp := range filePaths {
		fi, err := os.Lstat(fp)
		if err != nil || !fi.Mode().IsRegular() {
			continue
		}

		languages, _ := processor.DetectLanguage(fi.Name())
		if len(languages) == 0 {
			continue
		}
		lang := languages[0]

		content, err := os.ReadFile(fp)
		if err != nil {
			continue
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(fp string, lang string, content []byte) {
			defer func() { <-sem; wg.Done() }()
			job := &processor.FileJob{
				Filename: filepath.Base(fp),
				Language: lang,
				Content:  content,
				Bytes:    int64(len(content)),
			}
			processor.CountStats(job)

			relPath, _ := filepath.Rel(cleanPath, fp)
			mu.Lock()
			results = append(results, fileResult{
				language:   lang,
				filename:   filepath.Base(fp),
				location:   filepath.ToSlash(relPath),
				lines:      job.Lines,
				code:       job.Code,
				comment:    job.Comment,
				blank:      job.Blank,
				complexity: job.Complexity,
			})
			mu.Unlock()
		}(fp, lang, content)
	}

	wg.Wait()
	return results
}

func aggregateResults(results []fileResult) ([]SCCLanguageStat, []SCCFileStat) {
	langMap := make(map[string]*SCCLanguageStat)
	for _, r := range results {
		s, ok := langMap[r.language]
		if !ok {
			s = &SCCLanguageStat{Language: r.language}
			langMap[r.language] = s
		}
		s.Files++
		s.Lines += r.lines
		s.Code += r.code
		s.Comments += r.comment
		s.Blanks += r.blank
		s.Complexity += r.complexity
	}

	var langStats []SCCLanguageStat
	for _, s := range langMap {
		langStats = append(langStats, *s)
	}
	sort.Slice(langStats, func(i, j int) bool {
		return langStats[i].Code > langStats[j].Code
	})

	var fileStats []SCCFileStat
	for _, r := range results {
		fileStats = append(fileStats, SCCFileStat{
			Filename:   r.filename,
			Language:   r.language,
			Location:   r.location,
			Code:       r.code,
			Complexity: r.complexity,
		})
	}

	return langStats, fileStats
}

// FormatSummaryForWiki 生成 reference.md 中的代码概览
func FormatSummaryForWiki(langStats []SCCLanguageStat) string {
	if len(langStats) == 0 {
		return ""
	}

	// 过滤掉代码行数为 0 的语言
	var filtered []SCCLanguageStat
	var totalFiles, totalLines, totalCode, totalComments, totalBlanks, totalComplexity int64
	for _, s := range langStats {
		if s.Code <= 0 {
			continue
		}
		filtered = append(filtered, s)
		totalFiles += s.Files
		totalLines += s.Lines
		totalCode += s.Code
		totalComments += s.Comments
		totalBlanks += s.Blanks
		totalComplexity += s.Complexity
	}
	if len(filtered) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("| 语言 | 文件数 | 总行数 | 代码行 | 注释行 | 空行 | 复杂度 |\n")
	sb.WriteString("|------|--------|--------|--------|--------|------|--------|\n")
	for _, s := range filtered {
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d | %d | %d |\n",
			s.Language, s.Files, s.Lines, s.Code, s.Comments, s.Blanks, s.Complexity))
	}
	sb.WriteString(fmt.Sprintf("| **合计** | **%d** | **%d** | **%d** | **%d** | **%d** | **%d** |\n",
		totalFiles, totalLines, totalCode, totalComments, totalBlanks, totalComplexity))
	return sb.String()
}

// FormatTopFilesForWiki 生成 scc.md，包含语言汇总和 top 文件排名
func FormatTopFilesForWiki(langStats []SCCLanguageStat, files []SCCFileStat, topN int) string {
	if len(files) == 0 {
		return ""
	}

	var sb strings.Builder

	// 语言汇总
	sb.WriteString("# 代码统计\n\n")
	sb.WriteString(FormatSummaryForWiki(langStats))
	sb.WriteString("\n")

	// Top 文件排名
	sorted := make([]SCCFileStat, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Code > sorted[j].Code
	})

	if len(sorted) > topN {
		sorted = sorted[:topN]
	}

	sb.WriteString("## Top 文件排名（按代码行数）\n\n")
	sb.WriteString("| # | 文件 | 语言 | 代码行 | 复杂度 |\n")
	sb.WriteString("|---|------|------|--------|--------|\n")
	for i, f := range sorted {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %d | %d |\n",
			i+1, f.Filename, f.Language, f.Code, f.Complexity))
	}
	return sb.String()
}

// FormatSummaryForCLI 生成终端输出的汇总表格
func FormatSummaryForCLI(langStats []SCCLanguageStat) string {
	if len(langStats) == 0 {
		return ""
	}

	var totalFiles, totalCode, totalComplexity int64
	for _, s := range langStats {
		totalFiles += s.Files
		totalCode += s.Code
		totalComplexity += s.Complexity
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignCenter},
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
	})
	t.AppendHeader(table.Row{"语言", "文件数", "代码行", "复杂度"})
	for _, l := range langStats {
		t.AppendRow(table.Row{l.Language, l.Files, l.Code, l.Complexity})
	}
	t.AppendSeparator()
	t.AppendRow(table.Row{"合计", totalFiles, totalCode, totalComplexity})
	return t.Render()
}

// FormatTopFilesForCLI 生成终端输出的 top 文件表格
func FormatTopFilesForCLI(files []SCCFileStat, topN int) string {
	if len(files) == 0 {
		return ""
	}

	sorted := make([]SCCFileStat, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Code > sorted[j].Code
	})

	if len(sorted) > topN {
		sorted = sorted[:topN]
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
	})
	t.AppendHeader(table.Row{"#", "文件", "语言", "代码行", "复杂度"})
	for i, f := range sorted {
		t.AppendRow(table.Row{i + 1, f.Filename, f.Language, f.Code, f.Complexity})
	}
	return t.Render()
}
