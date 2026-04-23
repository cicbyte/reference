package repo

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var knownPlatforms = map[string]bool{
	"github.com":    true,
	"gitlab.com":    true,
	"bitbucket.org": true,
	"gitee.com":     true,
	"codeberg.org":  true,
	"sourcehut.org": true,
	"git.sr.ht":     true,
}

var platformShortNames = map[string]string{
	"github.com":    "github",
	"gitlab.com":    "gitlab",
	"gitee.com":     "gitee",
	"bitbucket.org": "bitbucket",
	"codeberg.org":  "codeberg",
	"sourcehut.org": "sourcehut",
	"git.sr.ht":     "sourcehut",
}

func PlatformShortName(host string) string {
	if s, ok := platformShortNames[host]; ok {
		return s
	}
	return host
}

type GitURLInfo struct {
	OriginalURL string
	Host        string
	Namespace   string
	RepoName    string
	IsKnown     bool
	CachePath   string // host/namespace/repo 或 other/sanitized
	LinkName    string // host-namespace-repo
	WikiSubPath string // platform/namespace/repo（用于 wiki 嵌套目录）
}

var (
	sshPattern     = regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)
	httpsPattern   = regexp.MustCompile(`^https?://([^/]+)/(.+?)(?:\.git)?$`)
)

func ParseGitURL(rawURL string, reposBase string) (*GitURLInfo, error) {
	rawURL = strings.TrimSpace(rawURL)

	var host, repoPath string

	if m := sshPattern.FindStringSubmatch(rawURL); m != nil {
		host = m[1]
		repoPath = m[2]
	} else if m := httpsPattern.FindStringSubmatch(rawURL); m != nil {
		host = m[1]
		repoPath = m[2]
	} else {
		if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
			host = u.Host
			repoPath = strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), ".git")
		} else {
			return nil, fmt.Errorf("无法解析 Git URL: %s", rawURL)
		}
	}

	parts := strings.Split(repoPath, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("仓库路径格式不正确（需要 namespace/repo）: %s", repoPath)
	}

	namespace := parts[0]
	repoName := parts[len(parts)-1]
	isKnown := knownPlatforms[host]

	info := &GitURLInfo{
		OriginalURL: rawURL,
		Host:        host,
		Namespace:   namespace,
		RepoName:    repoName,
		IsKnown:     isKnown,
	}

	if isKnown {
		shortHost := PlatformShortName(host)
		info.CachePath = filepath.Join(reposBase, shortHost, namespace, repoName)
		info.LinkName = fmt.Sprintf("%s-%s-%s", shortHost, namespace, repoName)
		info.WikiSubPath = shortHost + "/" + namespace + "/" + repoName
	} else {
		sanitized := strings.ReplaceAll(strings.ReplaceAll(host+"/"+repoPath, "/", "-"), ":", "-")
		info.CachePath = filepath.Join(reposBase, "other", sanitized)
		info.LinkName = sanitized
		info.WikiSubPath = PlatformShortName(host) + "/" + repoPath
	}

	return info, nil
}

func NormalizeGitURL(input string) string {
	input = strings.TrimSpace(input)
	if sshPattern.MatchString(input) || httpsPattern.MatchString(input) {
		return input
	}
	if strings.Contains(input, "/") && !strings.Contains(input, " ") {
		if !strings.HasPrefix(input, "http") && !strings.HasPrefix(input, "git@") {
			return "https://github.com/" + input
		}
	}
	return input
}
