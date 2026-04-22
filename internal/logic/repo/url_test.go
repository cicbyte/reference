package repo

import (
	"path/filepath"
	"testing"
)

func TestParseGitURL_HTTPS(t *testing.T) {
	reposDir := filepath.Join(t.TempDir(), "repos")
	info, err := ParseGitURL("https://github.com/cicbyte/memos-cli.git", reposDir)
	if err != nil {
		t.Fatal(err)
	}
	assert(t, "https://github.com/cicbyte/memos-cli.git", info.OriginalURL)
	assert(t, "github.com", info.Host)
	assert(t, "cicbyte", info.Namespace)
	assert(t, "memos-cli", info.RepoName)
	assert(t, true, info.IsKnown)
	assert(t, "github/cicbyte/memos-cli", info.WikiSubPath)
	assert(t, "github.com-cicbyte-memos-cli", info.LinkName)
}

func TestParseGitURL_SSH(t *testing.T) {
	reposDir := filepath.Join(t.TempDir(), "repos")
	info, err := ParseGitURL("git@gitlab.com:group/repo.git", reposDir)
	if err != nil {
		t.Fatal(err)
	}
	assert(t, "gitlab.com", info.Host)
	assert(t, "group", info.Namespace)
	assert(t, "repo", info.RepoName)
	assert(t, true, info.IsKnown)
	assert(t, "gitlab/group/repo", info.WikiSubPath)
}

func TestParseGitURL_Short(t *testing.T) {
	reposDir := filepath.Join(t.TempDir(), "repos")
	info, err := ParseGitURL("https://github.com/spf13/cobra", reposDir)
	if err != nil {
		t.Fatal(err)
	}
	assert(t, "cobra", info.RepoName)
	assert(t, "spf13/cobra", info.Namespace + "/" + info.RepoName)
}

func TestParseGitURL_UnknownPlatform(t *testing.T) {
	reposDir := filepath.Join(t.TempDir(), "repos")
	info, err := ParseGitURL("https://gitea.example.com/team/project", reposDir)
	if err != nil {
		t.Fatal(err)
	}
	assert(t, "gitea.example.com", info.Host)
	assert(t, false, info.IsKnown)
	assert(t, "gitea.example.com", PlatformShortName(info.Host)) // unknown → host itself
}

func TestParseGitURL_Invalid(t *testing.T) {
	_, err := ParseGitURL("not-a-url", filepath.Join(t.TempDir(), "repos"))
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/a/b", "https://github.com/a/b"},
		{"git@github.com:a/b.git", "git@github.com:a/b.git"},
		{"owner/repo", "https://github.com/owner/repo"},
		{"cicbyte/memos-cli", "https://github.com/cicbyte/memos-cli"},
	}
	for _, tc := range tests {
		got := NormalizeGitURL(tc.input)
		if got != tc.want {
			t.Errorf("NormalizeGitURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestPlatformShortName(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"github.com", "github"},
		{"gitlab.com", "gitlab"},
		{"gitee.com", "gitee"},
		{"bitbucket.org", "bitbucket"},
		{"codeberg.org", "codeberg"},
		{"sourcehut.org", "sourcehut"},
		{"git.sr.ht", "sourcehut"},
		{"gitea.example.com", "gitea.example.com"},
	}
	for _, tc := range tests {
		got := PlatformShortName(tc.host)
		if got != tc.want {
			t.Errorf("PlatformShortName(%q) = %q, want %q", tc.host, got, tc.want)
		}
	}
}

func assert(t *testing.T, want, got interface{}) {
	t.Helper()
	if want != got {
		t.Errorf("got %v, want %v", got, want)
	}
}
