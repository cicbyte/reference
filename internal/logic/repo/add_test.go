package repo

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cicbyte/reference/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestShortHash(t *testing.T) {
	h1 := shortHash("hello")
	h2 := shortHash("hello")
	h3 := shortHash("world")
	if h1 != h2 {
		t.Errorf("same input should produce same hash")
	}
	if h1 == h3 {
		t.Errorf("different input should produce different hash")
	}
	if len(h1) != 4 {
		t.Errorf("shortHash length = %d, want 4", len(h1))
	}
}

func TestFormatAddResult(t *testing.T) {
	tests := []struct {
		name string
		r    *AddResult
		d    time.Duration
		want string
	}{
		{"remote", &AddResult{RefName: "memos-cli", RefType: models.RefTypeRemote}, 5*time.Second, "[远程] 引用 'memos-cli' 添加成功 (5s)"},
		{"local", &AddResult{RefName: "my-lib", RefType: models.RefTypeLocal}, 123*time.Millisecond, "[本地] 引用 'my-lib' 添加成功 (123ms)"},
	}
	for _, tc := range tests {
		got := FormatAddResult(tc.r, tc.d)
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestResolveRemoteRefName_NoConflict(t *testing.T) {
	db := setupTestDB(t)
	defer os.RemoveAll(t.TempDir())
	indexer := NewRepoIndexer(db)
	tmpDir := t.TempDir()

	name := resolveRemoteRefName(tmpDir, indexer, "memos-cli", "cicbyte", "github.com", "github.com-cicbyte-memos-cli")
	if name != "memos-cli" {
		t.Errorf("got %q, want %q", name, "memos-cli")
	}
}

func TestResolveRemoteRefName_NamespaceConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable("repos")
	indexer := NewRepoIndexer(db)
	tmpDir := t.TempDir()

	// add existing repo with ref_name "memos-cli"
	indexer.Add(&models.Repo{
		ProjectDir: tmpDir, LinkName: "github.com-other-memos-cli", RefName: "memos-cli",
		RefType: models.RefTypeRemote,
	})

	name := resolveRemoteRefName(tmpDir, indexer, "memos-cli", "cicbyte", "github.com", "github.com-cicbyte-memos-cli")
	if name != "cicbyte-memos-cli" {
		t.Errorf("got %q, want %q (should fall to namespace-repo)", name, "cicbyte-memos-cli")
	}
}

func TestResolveRemoteRefName_FullConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable("repos")
	indexer := NewRepoIndexer(db)
	tmpDir := t.TempDir()

	indexer.Add(&models.Repo{
		ProjectDir: tmpDir, LinkName: "github.com-other-memos-cli", RefName: "memos-cli",
		RefType: models.RefTypeRemote,
	})
	indexer.Add(&models.Repo{
		ProjectDir: tmpDir, LinkName: "github.com-another-memos-cli", RefName: "cicbyte-memos-cli",
		RefType: models.RefTypeRemote,
	})

	name := resolveRemoteRefName(tmpDir, indexer, "memos-cli", "cicbyte", "github.com", "github.com-cicbyte-memos-cli")
	if name != "github.com-cicbyte-memos-cli" {
		t.Errorf("got %q, want %q (should fall to fullLinkName)", name, "github.com-cicbyte-memos-cli")
	}
}

func TestResolveRemoteRefName_DifferentProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable("repos")
	indexer := NewRepoIndexer(db)
	tmpDir := t.TempDir()

	// same ref_name in different project should not conflict
	otherDir := filepath.Join(tmpDir, "other")
	indexer.Add(&models.Repo{
		ProjectDir: otherDir, LinkName: "github.com-cicbyte-memos-cli", RefName: "memos-cli",
		RefType: models.RefTypeRemote,
	})

	name := resolveRemoteRefName(tmpDir, indexer, "memos-cli", "cicbyte", "github.com", "github.com-cicbyte-memos-cli")
	if name != "memos-cli" {
		t.Errorf("got %q, want %q (different project should not conflict)", name, "memos-cli")
	}
}

func TestResolveLocalRefName_NoConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable("repos")
	indexer := NewRepoIndexer(db)
	tmpDir := t.TempDir()

	name := resolveLocalRefName(tmpDir, indexer, "my-lib", "local-ab12-my-lib")
	if name != "my-lib" {
		t.Errorf("got %q, want %q", name, "my-lib")
	}
}

func TestResolveLocalRefName_Conflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Migrator().DropTable("repos")
	indexer := NewRepoIndexer(db)
	tmpDir := t.TempDir()

	indexer.Add(&models.Repo{
		ProjectDir: tmpDir, LinkName: "local-cd34-my-lib", RefName: "my-lib",
		RefType: models.RefTypeLocal,
	})

	name := resolveLocalRefName(tmpDir, indexer, "my-lib", "local-ab12-my-lib")
	if name != "local-ab12-my-lib" {
		t.Errorf("got %q, want %q", name, "local-ab12-my-lib")
	}
}

func TestResolveLocalWikiSubPath_NoConflict(t *testing.T) {
	wikiBase := t.TempDir()
	subPath := resolveLocalWikiSubPath(wikiBase, filepath.Join("home", "user", "project-a"))
	if subPath != "local/project-a" {
		t.Errorf("got %q, want %q", subPath, "local/project-a")
	}
}

func TestResolveLocalWikiSubPath_Conflict(t *testing.T) {
	wikiBase := t.TempDir()
	os.MkdirAll(filepath.Join(wikiBase, "local", "project-a"), 0755)

	subPath := resolveLocalWikiSubPath(wikiBase, filepath.Join("home", "user", "project-a"))
	if subPath == "local/project-a" {
		t.Errorf("expected hashed path on conflict, got %q", subPath)
	}
	if subPath[:6] != "local/" {
		t.Errorf("expected local/ prefix, got %q", subPath)
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Repo{}); err != nil {
		t.Fatal(err)
	}
	return db
}
