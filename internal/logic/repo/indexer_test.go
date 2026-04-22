package repo

import (
	"os"
	"testing"

	"github.com/cicbyte/reference/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Repo{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Migrator().DropTable("repos") })
	return db
}

func TestIndexer_AddAndGet(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	repo := &models.Repo{
		ProjectDir: "/project", LinkName: "github.com-a-b", RefName: "b",
		RefType: models.RefTypeRemote, RemoteURL: "https://github.com/a/b",
	}
	if err := idx.Add(repo); err != nil {
		t.Fatal(err)
	}

	got, err := idx.Get("/project", "github.com-a-b")
	if err != nil {
		t.Fatal(err)
	}
	if got.RefName != "b" {
		t.Errorf("RefName = %q, want %q", got.RefName, "b")
	}
	if got.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestIndexer_Upsert(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "a-b", RefName: "b", RefType: models.RefTypeRemote})
	// 用 Get 拿到已有记录，修改后再保存（模拟实际更新流程）
	existing, err := idx.Get("/p", "a-b")
	if err != nil {
		t.Fatal(err)
	}
	existing.Branch = "dev"
	if err := idx.Add(existing); err != nil {
		t.Fatal(err)
	}

	got, _ := idx.Get("/p", "a-b")
	if got.Branch != "dev" {
		t.Errorf("Branch = %q, want %q (upsert should update)", got.Branch, "dev")
	}

	var count int64
	db.Model(&models.Repo{}).Count(&count)
	if count != 1 {
		t.Errorf("count = %d, want 1 (upsert should not duplicate)", count)
	}
}

func TestIndexer_Remove(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "a-b", RefType: models.RefTypeRemote})
	if err := idx.Remove("/p", "a-b"); err != nil {
		t.Fatal(err)
	}

	_, err := idx.Get("/p", "a-b")
	if err == nil {
		t.Error("expected error after remove")
	}
}

func TestIndexer_List(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "a-b", RefType: models.RefTypeRemote})
	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "c-d", RefType: models.RefTypeRemote})
	idx.Add(&models.Repo{ProjectDir: "/other", LinkName: "e-f", RefType: models.RefTypeRemote})

	repos, err := idx.List("/p")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Errorf("len = %d, want 2", len(repos))
	}
}

func TestIndexer_GetByRefName(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "github.com-a-b", RefName: "b", RefType: models.RefTypeRemote})

	got, err := idx.GetByRefName("/p", "b")
	if err != nil {
		t.Fatal(err)
	}
	if got.LinkName != "github.com-a-b" {
		t.Errorf("LinkName = %q, want %q", got.LinkName, "github.com-a-b")
	}

	_, err = idx.GetByRefName("/p", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent ref_name")
	}
}

func TestIndexer_ListByType(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "a-b", RefType: models.RefTypeRemote})
	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "local-c", RefType: models.RefTypeLocal})

	remotes, _ := idx.ListByType("/p", models.RefTypeRemote)
	locals, _ := idx.ListByType("/p", models.RefTypeLocal)

	if len(remotes) != 1 || len(locals) != 1 {
		t.Errorf("remotes=%d locals=%d, want 1 1", len(remotes), len(locals))
	}
}

func TestIndexer_UniqueConstraint(t *testing.T) {
	db := newTestDB(t)
	idx := NewRepoIndexer(db)

	idx.Add(&models.Repo{ProjectDir: "/p", LinkName: "a-b", RefType: models.RefTypeRemote})
	idx.Add(&models.Repo{ProjectDir: "/p2", LinkName: "a-b", RefType: models.RefTypeRemote})

	var count int64
	db.Model(&models.Repo{}).Count(&count)
	if count != 2 {
		t.Errorf("count = %d, want 2 (same link_name different project_dir is allowed)", count)
	}
}

// suppress unused import warning
var _ = os.Stat
