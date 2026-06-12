// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func listFilesContext(t *testing.T, path string, extraQuery string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	target := "/files?path=" + url.QueryEscape(path)
	if extraQuery != "" {
		target += "&" + extraQuery
	}
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	ListFiles(ctx)
	return recorder
}

func decodeFileInfos(t *testing.T, recorder *httptest.ResponseRecorder) []FileInfo {
	t.Helper()
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var infos []FileInfo
	if err := json.Unmarshal(recorder.Body.Bytes(), &infos); err != nil {
		t.Fatalf("decoding response: %v (body: %s)", err, recorder.Body.String())
	}
	return infos
}

func fileNames(infos []FileInfo) []string {
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
	}
	sort.Strings(names)
	return names
}

func createListFilesTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "top.txt"), []byte("top"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "sub", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "child.txt"), []byte("child"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "nested", "deep.txt"), []byte("deep"), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestListFilesDepth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("default depth lists only direct entries", func(t *testing.T) {
		root := createListFilesTree(t)

		infos := decodeFileInfos(t, listFilesContext(t, root, ""))

		want := []string{"sub", "top.txt"}
		if got := fileNames(infos); !equalStrings(got, want) {
			t.Errorf("names = %v, want %v", got, want)
		}
	})

	t.Run("depth=2 includes children one level down", func(t *testing.T) {
		root := createListFilesTree(t)

		infos := decodeFileInfos(t, listFilesContext(t, root, "depth=2"))

		want := []string{"child.txt", "nested", "sub", "top.txt"}
		if got := fileNames(infos); !equalStrings(got, want) {
			t.Errorf("names = %v, want %v", got, want)
		}
		for _, info := range infos {
			if !strings.HasPrefix(info.Path, root+string(os.PathSeparator)) {
				t.Errorf("path %q not under root %q", info.Path, root)
			}
		}
	})

	t.Run("depth below 1 is rejected", func(t *testing.T) {
		root := createListFilesTree(t)

		recorder := listFilesContext(t, root, "depth=0")

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("non-integer depth is rejected", func(t *testing.T) {
		root := createListFilesTree(t)

		recorder := listFilesContext(t, root, "depth=1.5")

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})
}

func TestListFilesSymlinkRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on windows")
	}
	gin.SetMode(gin.TestMode)

	newSymlinkRoot := func(t *testing.T) (string, string) {
		t.Helper()
		root := createListFilesTree(t)
		link := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(root, link); err != nil {
			t.Fatal(err)
		}
		return root, link
	}

	t.Run("recursive listing matches the real root", func(t *testing.T) {
		root, link := newSymlinkRoot(t)

		realInfos := decodeFileInfos(t, listFilesContext(t, root, "depth=2"))
		linkInfos := decodeFileInfos(t, listFilesContext(t, link, "depth=2"))

		if got, want := fileNames(linkInfos), fileNames(realInfos); !equalStrings(got, want) {
			t.Errorf("symlink root names = %v, want %v", got, want)
		}
	})

	t.Run("recursive listing reports paths under the symlink root", func(t *testing.T) {
		_, link := newSymlinkRoot(t)

		infos := decodeFileInfos(t, listFilesContext(t, link, "depth=2"))

		if len(infos) == 0 {
			t.Fatal("expected entries for symlinked root, got none")
		}
		for _, info := range infos {
			if !strings.HasPrefix(info.Path, link+string(os.PathSeparator)) {
				t.Errorf("path %q not under symlink root %q", info.Path, link)
			}
		}
	})

	t.Run("shallow listing of symlink root is unchanged", func(t *testing.T) {
		_, link := newSymlinkRoot(t)

		infos := decodeFileInfos(t, listFilesContext(t, link, ""))

		want := []string{"sub", "top.txt"}
		if got := fileNames(infos); !equalStrings(got, want) {
			t.Errorf("names = %v, want %v", got, want)
		}
	})
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
