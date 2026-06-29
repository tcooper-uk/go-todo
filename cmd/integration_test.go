package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var todoBin string

// TestMain builds the binary once for all integration tests.
func TestMain(m *testing.M) {
	bin, err := os.CreateTemp("", "todo-integration-*")
	if err != nil {
		panic(err)
	}
	bin.Close()
	todoBin = bin.Name()

	if out, err := exec.Command("go", "build", "-o", todoBin, ".").CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	code := m.Run()
	os.Remove(todoBin)
	os.Exit(code)
}

// run invokes the todo binary with the given args, using dir as HOME so
// storage is isolated to a temp directory. Returns stdout, stderr, and whether
// the command exited successfully.
func run(t *testing.T, homeDir string, args ...string) (stdout, stderr string, ok bool) {
	t.Helper()
	cmd := exec.Command(todoBin, args...)
	cmd.Env = append(os.Environ(), "HOME="+homeDir, "TODO_BACKEND=sqlite")

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err == nil
}

// tempHome creates a temporary directory to use as HOME for one test.
func tempHome(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "todo-home-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	if err := os.MkdirAll(filepath.Join(dir, ".todo"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// mustRun fails the test immediately if the command exits non-zero.
func mustRun(t *testing.T, home string, args ...string) string {
	t.Helper()
	stdout, stderr, ok := run(t, home, args...)
	if !ok {
		t.Fatalf("%v failed\nstderr: %s", args, stderr)
	}
	return stdout
}

// --- add ---

func TestAdd_BasicItem(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Buy oat milk")
	out := mustRun(t, home, "list")
	if !strings.Contains(out, "Buy oat milk") {
		t.Errorf("expected item in list output, got:\n%s", out)
	}
}

func TestAdd_WithPriority(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--priority", "high", "Urgent task")
	out := mustRun(t, home, "list")
	if !strings.Contains(out, "[H]") {
		t.Errorf("expected high priority marker [H] in output, got:\n%s", out)
	}
}

func TestAdd_WithDueDate(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--due", "2099-12-31", "Future task")
	out := mustRun(t, home, "list")
	if !strings.Contains(out, "31 Dec 99") {
		t.Errorf("expected due date in output, got:\n%s", out)
	}
}

func TestAdd_WithTag(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--tag", "work", "Tagged task")
	out := mustRun(t, home, "list")
	if !strings.Contains(out, "#work") {
		t.Errorf("expected tag #work in output, got:\n%s", out)
	}
}

// --- list ---

func TestList_HidesDoneByDefault(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Open task")
	mustRun(t, home, "add", "Done task")
	mustRun(t, home, "done", "2")

	out := mustRun(t, home, "list")
	if strings.Contains(out, "Done task") {
		t.Errorf("default list should hide done items, got:\n%s", out)
	}
	if !strings.Contains(out, "Open task") {
		t.Errorf("default list should show open items, got:\n%s", out)
	}
}

func TestList_All_IncludesDone(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Open task")
	mustRun(t, home, "add", "Done task")
	mustRun(t, home, "done", "2")

	out := mustRun(t, home, "list", "--all")
	if !strings.Contains(out, "Done task") {
		t.Errorf("--all should include done items, got:\n%s", out)
	}
}

func TestList_OnlyDone(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Open task")
	mustRun(t, home, "add", "Done task")
	mustRun(t, home, "done", "2")

	out := mustRun(t, home, "list", "--done")
	if strings.Contains(out, "Open task") {
		t.Errorf("--done should exclude open items, got:\n%s", out)
	}
	if !strings.Contains(out, "Done task") {
		t.Errorf("--done should include done items, got:\n%s", out)
	}
}

func TestList_FilterByPriority(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--priority", "high", "High task")
	mustRun(t, home, "add", "--priority", "low", "Low task")

	out := mustRun(t, home, "list", "--priority", "high")
	if !strings.Contains(out, "High task") {
		t.Errorf("expected high priority item, got:\n%s", out)
	}
	if strings.Contains(out, "Low task") {
		t.Errorf("expected low priority item to be filtered out, got:\n%s", out)
	}
}

func TestList_FilterByTag(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--tag", "work", "Work task")
	mustRun(t, home, "add", "--tag", "home", "Home task")

	out := mustRun(t, home, "list", "--tag", "work")
	if !strings.Contains(out, "Work task") {
		t.Errorf("expected work-tagged item, got:\n%s", out)
	}
	if strings.Contains(out, "Home task") {
		t.Errorf("expected home-tagged item to be filtered out, got:\n%s", out)
	}
}

func TestList_Overdue(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--due", "2020-01-01", "Overdue task")
	mustRun(t, home, "add", "--due", "2099-12-31", "Future task")

	out := mustRun(t, home, "list", "--overdue")
	if !strings.Contains(out, "Overdue task") {
		t.Errorf("expected overdue item, got:\n%s", out)
	}
	if strings.Contains(out, "Future task") {
		t.Errorf("expected future item to be excluded from --overdue, got:\n%s", out)
	}
}

// --- edit ---

func TestEdit_Rename(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Old name")
	mustRun(t, home, "edit", "1", "New name")

	out := mustRun(t, home, "list")
	if strings.Contains(out, "Old name") {
		t.Errorf("old name should be gone after rename, got:\n%s", out)
	}
	if !strings.Contains(out, "New name") {
		t.Errorf("expected new name in list, got:\n%s", out)
	}
}

func TestEdit_Priority(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Some task")
	mustRun(t, home, "edit", "1", "--priority", "medium")

	out := mustRun(t, home, "list")
	if !strings.Contains(out, "[M]") {
		t.Errorf("expected medium priority marker after edit, got:\n%s", out)
	}
}

func TestEdit_DueDate(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Some task")
	mustRun(t, home, "edit", "1", "--due", "2099-06-15")

	out := mustRun(t, home, "list")
	if !strings.Contains(out, "15 Jun 99") {
		t.Errorf("expected due date after edit, got:\n%s", out)
	}
}

func TestEdit_ClearDueDate(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--due", "2099-06-15", "Some task")
	mustRun(t, home, "edit", "1", "--due", "-")

	// Use the detail view — it only prints "Due:" when a date is set.
	out := mustRun(t, home, "1")
	if strings.Contains(out, "Due:") {
		t.Errorf("due date should be cleared, got:\n%s", out)
	}
}

func TestEdit_Tags(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "--tag", "old", "Some task")
	mustRun(t, home, "edit", "1", "--tag", "new")

	out := mustRun(t, home, "list")
	if strings.Contains(out, "#old") {
		t.Errorf("old tag should be replaced, got:\n%s", out)
	}
	if !strings.Contains(out, "#new") {
		t.Errorf("expected new tag after edit, got:\n%s", out)
	}
}

func TestEdit_Done(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Some task")
	mustRun(t, home, "edit", "1", "--done", "true")

	out := mustRun(t, home, "list", "--done")
	if !strings.Contains(out, "[x]") {
		t.Errorf("expected done marker [x] after edit --done true, got:\n%s", out)
	}
}

// --- done / reopen ---

func TestDone_MarksComplete(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Finish me")
	mustRun(t, home, "done", "1")

	// Should be hidden from default list.
	out := mustRun(t, home, "list")
	if strings.Contains(out, "Finish me") {
		t.Errorf("done item should be hidden from default list, got:\n%s", out)
	}

	// Should appear with --all and show [x].
	out = mustRun(t, home, "list", "--all")
	if !strings.Contains(out, "[x]") {
		t.Errorf("expected done marker [x] in --all list, got:\n%s", out)
	}
}

func TestReopen_MarksOpen(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Reopen me")
	mustRun(t, home, "done", "1")
	mustRun(t, home, "reopen", "1")

	out := mustRun(t, home, "list")
	if !strings.Contains(out, "Reopen me") {
		t.Errorf("reopened item should appear in default list, got:\n%s", out)
	}
	if !strings.Contains(out, "[ ]") {
		t.Errorf("expected open marker [ ] after reopen, got:\n%s", out)
	}
}

// --- delete ---

func TestDelete_Single(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Delete me")
	mustRun(t, home, "delete", "1")

	out := mustRun(t, home, "list")
	if strings.Contains(out, "Delete me") {
		t.Errorf("deleted item should not appear in list, got:\n%s", out)
	}
}

func TestDelete_Multiple(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "First")
	mustRun(t, home, "add", "Second")
	mustRun(t, home, "add", "Keep me")
	mustRun(t, home, "delete", "1", "2")

	out := mustRun(t, home, "list")
	if strings.Contains(out, "First") || strings.Contains(out, "Second") {
		t.Errorf("deleted items should not appear in list, got:\n%s", out)
	}
	if !strings.Contains(out, "Keep me") {
		t.Errorf("non-deleted item should remain, got:\n%s", out)
	}
}

// --- output format ---

func TestList_MultilineNameShowsFirstLineOnly(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "First line\nSecond line")

	out := mustRun(t, home, "list")
	if !strings.Contains(out, "First line") {
		t.Errorf("expected first line in list output, got:\n%s", out)
	}
	if strings.Contains(out, "Second line") {
		t.Errorf("second line should not appear in list output, got:\n%s", out)
	}
	if !strings.Contains(out, "…") {
		t.Errorf("expected ellipsis indicating more content, got:\n%s", out)
	}
}

func TestDetail_MultilineNameShowsFullText(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "First line\nSecond line")

	out := mustRun(t, home, "1")
	if !strings.Contains(out, "First line") {
		t.Errorf("expected first line in detail view, got:\n%s", out)
	}
	if !strings.Contains(out, "Second line") {
		t.Errorf("expected second line in detail view, got:\n%s", out)
	}
}

// --- error cases ---

func TestError_UnknownCommand(t *testing.T) {
	home := tempHome(t)
	_, _, ok := run(t, home, "notacommand")
	if ok {
		t.Error("expected non-zero exit for unknown command")
	}
}

func TestError_EditMissingID(t *testing.T) {
	home := tempHome(t)
	_, _, ok := run(t, home, "edit")
	if ok {
		t.Error("expected non-zero exit when edit is called with no ID")
	}
}

func TestError_DoneMissingID(t *testing.T) {
	home := tempHome(t)
	_, _, ok := run(t, home, "done")
	if ok {
		t.Error("expected non-zero exit when done is called with no ID")
	}
}

func TestError_InvalidDueDate(t *testing.T) {
	home := tempHome(t)
	mustRun(t, home, "add", "Some task")
	_, _, ok := run(t, home, "edit", "1", "--due", "not-a-date")
	if ok {
		t.Error("expected non-zero exit for invalid due date")
	}
}
