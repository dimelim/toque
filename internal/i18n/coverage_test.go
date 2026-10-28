package i18n

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var (
	repo        = filepath.Join("..", "..")
	goCall      = regexp.MustCompile("i18n\\.(?:T|Tf|Mark)\\(\\s*(\"(?:[^\"\\\\\\n]|\\\\.)*\"|`[^`]*`)")
	tsCall      = regexp.MustCompile("(?:^|[^\\w.$])(?:t|mark)\\(\\s*(\"(?:[^\"\\\\\\n]|\\\\.)*\"|'(?:[^'\\\\\\n]|\\\\.)*'|`(?:[^`\\\\]|\\\\.)*`)")
	tsDynamic   = regexp.MustCompile("(?:^|[^\\w.$])(?:t|mark)\\(\\s*`[^`]*\\$\\{")
	namedHole   = regexp.MustCompile(`\{(\w+)\}`)
	formatVerb  = regexp.MustCompile(`%(?:\[\d+\])?[-+# 0]*\d*(?:\.\d+)?[a-zA-Z%]`)
	targetLangs = []Lang{PT, EN}
)

type use struct {
	key      string
	where    string
	goFormat bool
}

func unquoteTS(lit string) (string, error) {
	switch lit[0] {
	case '`':
		return strings.ReplaceAll(lit[1:len(lit)-1], "\\`", "`"), nil
	case '\'':
		inner := strings.ReplaceAll(lit[1:len(lit)-1], "\\'", "'")
		return strconv.Unquote(`"` + strings.ReplaceAll(inner, `"`, `\"`) + `"`)
	}
	return strconv.Unquote(lit)
}

func walk(t *testing.T, dir string, keep func(string) bool, each func(path, text string)) {
	t.Helper()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if n := d.Name(); n == "node_modules" || n == "wailsjs" || n == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if !keep(p) {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		each(p, string(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func line(text string, at int) string { return strconv.Itoa(strings.Count(text[:at], "\n") + 1) }

func collect(t *testing.T) []use {
	var uses []use
	goFile := func(p string) bool { return strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") }
	scanGo := func(p, text string) {
		for _, m := range goCall.FindAllStringSubmatchIndex(text, -1) {
			key, err := strconv.Unquote(text[m[2]:m[3]])
			if err != nil {
				t.Errorf("%s:%s: %v", p, line(text, m[0]), err)
				continue
			}
			uses = append(uses, use{key, p + ":" + line(text, m[0]), true})
		}
	}
	walk(t, filepath.Join(repo, "internal"), goFile, scanGo)
	walk(t, filepath.Join(repo, "cmd"), goFile, scanGo)
	walk(t, repo, func(p string) bool { return filepath.Dir(p) == repo && goFile(p) }, scanGo)
	tsFile := func(p string) bool { return strings.HasSuffix(p, ".ts") || strings.HasSuffix(p, ".tsx") }
	walk(t, filepath.Join(repo, "frontend", "src"), tsFile, func(p, text string) {
		for _, m := range tsDynamic.FindAllStringIndex(text, -1) {
			t.Errorf("%s:%s: t() needs a fixed text, use {name} placeholders instead of ${}", p, line(text, m[0]))
		}
		for _, m := range tsCall.FindAllStringSubmatchIndex(text, -1) {
			lit := text[m[2]:m[3]]
			if strings.HasPrefix(lit, "`") && strings.Contains(lit, "${") {
				continue
			}
			key, err := unquoteTS(lit)
			if err != nil {
				t.Errorf("%s:%s: %v", p, line(text, m[0]), err)
				continue
			}
			uses = append(uses, use{key, p + ":" + line(text, m[0]), false})
		}
	})
	return uses
}

func dictionaries(t *testing.T) map[Lang]map[string]string {
	t.Helper()
	out := map[Lang]map[string]string{}
	paths, err := filepath.Glob(filepath.Join(repo, "frontend", "src", "locales", "*", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		l, ok := Parse(strings.TrimSuffix(filepath.Base(p), ".json"))
		if !ok {
			t.Errorf("%s: unknown language file", p)
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		entries := map[string]string{}
		if err := json.Unmarshal(b, &entries); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if out[l] == nil {
			out[l] = map[string]string{}
		}
		for k, v := range entries {
			if _, dup := out[l][k]; dup {
				t.Errorf("%s: %q is also in another %s file", p, k, l)
			}
			out[l][k] = v
		}
	}
	return out
}

func holes(s string, goFormat bool) []string {
	if goFormat {
		return formatVerb.FindAllString(s, -1)
	}
	var out []string
	for _, m := range namedHole.FindAllStringSubmatch(s, -1) {
		out = append(out, "{"+m[1]+"}")
	}
	sort.Strings(out)
	return out
}

func TestEveryTextIsTranslated(t *testing.T) {
	uses := collect(t)
	dicts := dictionaries(t)
	used := map[string]bool{}
	for _, u := range uses {
		used[u.key] = true
		for _, l := range targetLangs {
			v, ok := dicts[l][u.key]
			switch {
			case !ok || strings.TrimSpace(v) == "":
				t.Errorf("%s: missing %s translation for %q", u.where, l, u.key)
			case !slices.Equal(holes(u.key, u.goFormat), holes(v, u.goFormat)):
				t.Errorf("%s: %s translation of %q changes its placeholders: %q", u.where, l, u.key, v)
			}
		}
	}
	for _, l := range targetLangs {
		for k := range dicts[l] {
			if !used[k] {
				t.Errorf("%s dictionary has %q but no code uses it", l, k)
			}
		}
	}
}

func TestTranslate(t *testing.T) {
	if err := Load(os.DirFS(filepath.Join(repo, "frontend", "src", "locales"))); err != nil {
		t.Fatal(err)
	}
	defer Set(ES)
	Set(ES)
	if got := T("texto sin traducir"); got != "texto sin traducir" {
		t.Fatalf("ES should return the source text, got %q", got)
	}
	Set(PT)
	if got := T("texto sin traducir"); got != "texto sin traducir" {
		t.Fatalf("missing keys should fall back to the source text, got %q", got)
	}
}
