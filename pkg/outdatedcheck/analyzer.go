package outdatedcheck

import (
	"bytes"
	"fmt"
	"go/token"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/tools/go/analysis"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Analyzer = &analysis.Analyzer{
	Name: "outdatedcheck",
	Doc:  "git.mulk.net paketleri için güncel versiyon kontrolü yapar",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Projenin kök dizinini bul
	projectRoot, err := findGoModRoot(pass)
	if err != nil {
		return nil, err
	}

	goModPath := filepath.Join(projectRoot, "go.mod")

	data, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	modFile, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return nil, err
	}

	for _, req := range modFile.Require {
		if strings.HasPrefix(req.Mod.Path, "git.mulk.net") {
			latestVersion, err := getLatestVersion(req.Mod.Path)
			if err != nil {
				pass.Reportf(token.NoPos, "Modül '%s' için versiyon kontrolü yapılamadı: %v", req.Mod.Path, err)
				continue
			}

			currentVersion := req.Mod.Version
			if !semver.IsValid(currentVersion) {
				pass.Reportf(token.NoPos, "Modül '%s' geçersiz bir versiyona sahip: %s", req.Mod.Path, currentVersion)
				continue
			}

			if semver.Compare(latestVersion, currentVersion) > 0 {
				pass.Reportf(token.NoPos, "Modül '%s' eski bir versiyona sahip (%s), en son versiyon: %s", req.Mod.Path, currentVersion, latestVersion)
			}
		}
	}

	return nil, nil
}

func getLatestVersion(modulePath string) (string, error) {
	// 'git ls-remote --tags <modulePath>' komutunu çalıştır
	cmd := exec.Command("git", "ls-remote", "--tags", modulePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Hata çıktısını da yakala
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git komutu başarısız oldu: %v, çıktı: %s", err, out.String())
	}

	lines := strings.Split(out.String(), "\n")
	var versions []string
	for _, line := range lines {
		parts := strings.Split(line, "refs/tags/")
		if len(parts) == 2 {
			tag := strings.TrimSpace(parts[1])
			// Annotated tag'leri temizle (örn: ^{})
			tag = strings.TrimSuffix(tag, "^{}")
			// Eğer tag 'v' ile başlamıyorsa ekleyelim
			if !strings.HasPrefix(tag, "v") {
				tag = "v" + tag
			}
			if semver.IsValid(tag) {
				versions = append(versions, tag)
			}
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("tag bulunamadı")
	}

	// Semantik versiyon sıralaması yapın
	semver.Sort(versions)

	// En son versiyonu döndürün
	return versions[len(versions)-1], nil
}

func findGoModRoot(pass *analysis.Pass) (string, error) {
	// İlk dosyanın yolunu al
	if len(pass.Files) == 0 {
		return "", fmt.Errorf("analiz edilen dosya bulunamadı")
	}

	firstFile := pass.Fset.Position(pass.Files[0].Pos()).Filename
	if firstFile == "" {
		return "", fmt.Errorf("dosya yolu alınamadı")
	}

	// Dosyanın bulunduğu dizinden başlayarak go.mod dosyasını arayın
	dir := filepath.Dir(firstFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // Kök dizine ulaştık
		}
		dir = parentDir
	}

	return "", fmt.Errorf("go.mod dosyası bulunamadı")
}
