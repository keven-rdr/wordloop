// Command check-structure impoe os limites estruturais da API (AGENTS.md, regra 5):
// arquivo <= 300 linhas, <= 12 arquivos por pacote e <= 60 arquivos por modulo.
// Codigo gerado (gen/, db/) e arquivos de teste nao entram na conta de arquivos por pacote.
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxLines      = 300
	maxPerPackage = 12
	maxPerModule  = 60
)

var roots = []string{"internal", "cmd", "scripts"}

func scanAll() (pkgs, mods map[string]int, problems []string) {
	pkgs, mods = map[string]int{}, map[string]int{}
	for _, root := range roots {
		p, m, pr := scan(root)
		for k, v := range p {
			pkgs[k] += v
		}
		for k, v := range m {
			mods[k] += v
		}
		problems = append(problems, pr...)
	}
	return pkgs, mods, problems
}

func main() {
	pkgFiles, modFiles, problems := scanAll()
	problems = append(problems, over(pkgFiles, maxPerPackage, "pacote")...)
	problems = append(problems, over(modFiles, maxPerModule, "modulo")...)
	sort.Strings(problems)
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "check-structure: %d violacao(oes)\n", len(problems))
		os.Exit(1)
	}
	fmt.Println("check-structure: ok")
}

func generated(path string) bool {
	p := filepath.ToSlash(path)
	return strings.Contains(p, "/gen/") || strings.Contains(p, "/db/")
}

func scan(root string) (pkgs, mods map[string]int, problems []string) {
	pkgs, mods = map[string]int{}, map[string]int{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || generated(path) {
			return err
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		if n := bytes.Count(data, []byte("\n")) + 1; n > maxLines {
			problems = append(problems, fmt.Sprintf("%s: %d linhas (limite %d)", filepath.ToSlash(path), n, maxLines))
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		pkgs[filepath.ToSlash(filepath.Dir(path))]++
		mods[module(path)]++
		return nil
	})
	if err != nil {
		problems = append(problems, "erro ao varrer: "+err.Error())
	}
	return pkgs, mods, problems
}

// module e o primeiro diretorio depois de internal/ (ex.: internal/learning/app/x.go -> internal/learning).
func module(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) < 2 {
		return path
	}
	return strings.Join(parts[:2], "/")
}

func over(counts map[string]int, limit int, kind string) []string {
	var out []string
	for name, n := range counts {
		if n > limit {
			out = append(out, fmt.Sprintf("%s: %d arquivos no %s (limite %d)", name, n, kind, limit))
		}
	}
	return out
}
