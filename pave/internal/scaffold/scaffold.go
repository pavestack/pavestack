package scaffold

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pavestack/pave/internal/validate"
	"github.com/spf13/afero"
)

// WalkDir is a helper that walks the file tree rooted at root, calling fn for each file or directory
// in the tree, including root, using the provided afero.Fs.
func WalkDir(fsys afero.Fs, root string, fn fs.WalkDirFunc) error {
	return afero.Walk(fsys, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fn(path, nil, err)
		}
		return fn(path, fs.FileInfoToDirEntry(info), nil)
	})
}

// CreateService scaffolds a new internal API service.
func CreateService(fsys afero.Fs, repoRoot string, request validate.ServiceRequest) (string, error) {
	templateDir := filepath.Join(repoRoot, "service-template-api")
	serviceDir := filepath.Join(repoRoot, "services", request.Name+"-api")

	if err := copyDir(fsys, templateDir, serviceDir); err != nil {
		return "", fmt.Errorf("copy template: %w", err)
	}

	replacements := []string{
		"github.com/pavestack/service-template-api", fmt.Sprintf("github.com/pavestack/services/%s-api", request.Name),
		"pavestack/service-template-api", fmt.Sprintf("pavestack/%s-api", request.Name),
		"SERVICE_NAME: service-template-api", fmt.Sprintf("SERVICE_NAME: %s-api", request.Name),
		"service-template-api", request.Name + "-api",
		"team-platform", request.Team,
	}

	if err := walkReplace(fsys, serviceDir, replacements); err != nil {
		return "", err
	}

	if err := renameHelmChart(fsys, serviceDir, request.Name); err != nil {
		return "", err
	}

	if request.Database {
		if err := appendDatabaseStub(fsys, serviceDir); err != nil {
			return "", err
		}
	}

	metadataPath := filepath.Join(serviceDir, ".pavestack", "service-request.json")
	if err := writeServiceMetadata(fsys, metadataPath, request); err != nil {
		return "", err
	}

	return serviceDir, nil
}

func copyDir(fsys afero.Fs, src, dst string) error {
	return WalkDir(fsys, src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if entry.IsDir() {
			return fsys.MkdirAll(target, 0o755)
		}

		if shouldSkip(rel) {
			return nil
		}

		return copyFile(fsys, path, target)
	})
}

func shouldSkip(rel string) bool {
	skip := []string{".git"}
	for _, part := range skip {
		if strings.Contains(rel, part) {
			return true
		}
	}
	return false
}

func copyFile(fsys afero.Fs, src, dst string) error {
	if err := fsys.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := fsys.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := fsys.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func walkReplace(fsys afero.Fs, root string, replacements []string) error {
	return WalkDir(fsys, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		data, err := afero.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		content := string(data)
		replacer := strings.NewReplacer(replacements...)
		newContent := replacer.Replace(content)

		if newContent == content {
			return nil
		}
		return afero.WriteFile(fsys, path, []byte(newContent), 0o644)
	})
}

func renameHelmChart(fsys afero.Fs, serviceDir, name string) error {
	oldChart := filepath.Join(serviceDir, "deploy", "helm", "service-template-api")
	newChart := filepath.Join(serviceDir, "deploy", "helm", name+"-api")
	if err := fsys.Rename(oldChart, newChart); err != nil {
		return fmt.Errorf("rename helm chart: %w", err)
	}
	return nil
}

func appendDatabaseStub(fsys afero.Fs, serviceDir string) error {
	readme := filepath.Join(serviceDir, "README.md")
	content := "\n\n## Database\n\nThis service requested a managed database. Provision credentials via the platform secrets workflow.\n"
	f, err := fsys.OpenFile(readme, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func writeServiceMetadata(fsys afero.Fs, path string, request validate.ServiceRequest) error {
	if err := fsys.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload := fmt.Sprintf(`{
  "name": %q,
  "team": %q,
  "database": %t
}
`, request.Name, request.Team, request.Database)
	return afero.WriteFile(fsys, path, []byte(payload), 0o644)
}
