package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
)

// Module yapısı Go modüllerini temsil eder
type Module struct {
	Path    string
	Version string
	Update  *struct {
		Path    string
		Version string
	}
}

// CustomLinter yapısı lint arayüzünü uygulamak için kullanılır
type CustomLinter struct{}

// Run metodu bir linter işlevini uygular
func (c *CustomLinter) Run() error {
	// "go list -u -m -json all" komutunu çalıştırarak modül bilgilerini alın
	cmd := exec.Command("go", "list", "-u", "-m", "-json", "all")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("go list komutu çalıştırılırken hata oluştu: %v", err)
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

		// Modülleri kontrol et ve sonucu yazdır
		if mod.Update != nil {
			fmt.Fprintf(w, "%s\t%s\t%s\tGüncel Değil\n", mod.Path, mod.Version, mod.Update.Version)
		} else {
			fmt.Fprintf(w, "%s\t%s\t-\tGüncel\n", mod.Path, mod.Version)
		}
	}

	// Tablo yazıcıyı temizle
	w.Flush()

	return nil
}

// Plugin'i dışa aktar
var Plugin = &CustomLinter{}
