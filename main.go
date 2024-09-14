package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/golangci/golangci-lint/pkg/lint"
	"github.com/golangci/golangci-lint/pkg/result"
)

// Go modüllerini temsil eden Module yapısı
type Module struct {
	Path    string
	Version string
	Update  *struct {
		Path    string
		Version string
	}
}

// Konfigürasyonları tutan Config yapısı
type Config struct {
	ModulePrefix string `yaml:"module_prefix"`
}

// lint.Runner arayüzünü uygulamak için kullanılan CustomLinter yapısı
type CustomLinter struct {
	Config Config
}

// Run metodu lint.Runner arayüzünü uygular
func (c *CustomLinter) Run(ctx *lint.GoLintContext) ([]*result.Issue, error) {
	var issues []*result.Issue

	// Modül bilgilerini almak için "go list -u -m -json all" komutunu çalıştır
	cmd := exec.Command("go", "list", "-u", "-m", "-json", "all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list komutu çalıştırılırken hata oluştu: %v", err)
	}

	// Tablo yazıcıyı oluştur
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)

	// Tablo başlıkları
	fmt.Fprintln(w, "MODULE\tVERSION\tNEW VERSION\tSTATUS")

	// JSON formatında modülleri çözümle
	dec := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var mod Module
		if err := dec.Decode(&mod); err != nil {
			break
		}

		// Modülleri kontrol et; eğer önek (prefix) yapılandırılmışsa, sadece o önekle başlayan modülleri kontrol et, aksi halde hepsini kontrol et
		if c.Config.ModulePrefix == "" || strings.HasPrefix(mod.Path, c.Config.ModulePrefix) {
			if mod.Update != nil {
				fmt.Fprintf(w, "%s\t%s\t%s\tGüncel Değil\n", mod.Path, mod.Version, mod.Update.Version)
				issues = append(issues, &result.Issue{
					FromLinter: "MyCustomLinter",
					Text:       fmt.Sprintf("Modül %s güncel değil. Mevcut versiyon: %s, Yeni versiyon: %s", mod.Path, mod.Version, mod.Update.Version),
					Pos:        lint.Position{Filename: mod.Path, Line: 0},
					Severity:   lint.SeverityWarning,
				})
			} else {
				fmt.Fprintf(w, "%s\t%s\t-\tGüncel\n", mod.Path, mod.Version)
			}
		}
	}

	// Tablo yazıcıyı temizle
	w.Flush()

	return issues, nil
}

// loadConfig fonksiyonu ayarları ortam değişkenlerinden veya .golangci.yaml dosyasından yükler
func loadConfig() *CustomLinter {
	// Ortam değişkeninden (varsa) modül önekini yükle
	modulePrefix := os.Getenv("GOLANGCI_LINT_PLUGIN_PREFIX")

	// Yüklenen yapılandırma ile CustomLinter döndür
	return &CustomLinter{
		Config: Config{
			ModulePrefix: modulePrefix,
		},
	}
}

// Eklentiyi dışa aktar
var Plugin lint.Runner = loadConfig()
