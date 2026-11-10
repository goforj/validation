package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fixture creates an independent source tree so failure tests never edit the checkout.
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sources, err := filepath.Glob("../../*.go")
	if err != nil {
		t.Fatal(err)
	}
	examples, err := filepath.Glob("../../examples/*/main.go")
	if err != nil {
		t.Fatal(err)
	}
	sources = append(sources, examples...)
	sources = append(sources, "../../docs/README.tmpl")
	for _, source := range sources {
		relative, err := filepath.Rel("../..", source)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// write replaces a fixture file while keeping error handling out of each regression case.
func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestGenerate verifies deterministic projection and every generated navigation target.
func TestGenerate(t *testing.T) {
	root := fixture(t)
	t.Chdir(filepath.Join(root, "docs"))
	main()
	first, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := generate(root); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("generation drifted")
	}
	output := string(first)
	for _, match := range regexp.MustCompile(`\]\(#([^)]*)\)`).FindAllStringSubmatch(output, -1) {
		if !strings.Contains(output, `id="`+match[1]+`"`) {
			t.Fatalf("unresolved anchor: %s", match[1])
		}
	}
	for _, entry := range sections {
		heading := "### " + strings.Join(entry.apis, " · ") + "\n"
		if !strings.Contains(output, heading) {
			t.Fatal("missing API section", heading)
		}
		_, after, _ := strings.Cut(output, heading)
		section, _, _ := strings.Cut(after, "\n## ")
		if !strings.Contains(section, "```go") {
			t.Fatal("section has no example", heading)
		}
	}
}

// TestSourceErrors protects generation from missing and malformed authoritative inputs.
func TestSourceErrors(t *testing.T) {
	for _, name := range []string{"root", "source", "template", "program", "snippet", "syntax", "symbol", "currency_list", "description", "demonstration", "duplicate", "marker", "write"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			entries := sections
			switch name {
			case "root":
				root = filepath.Join(root, "absent")
			case "source":
				write(t, filepath.Join(root, "broken.go"), "package ???")
			case "template":
				if err := os.Remove(filepath.Join(root, "docs", "README.tmpl")); err != nil {
					t.Fatal(err)
				}
			case "program":
				entries = []section{{"Values", []string{"String"}, "missing", []string{"quickstart"}, ""}}
			case "snippet":
				write(t, filepath.Join(root, "examples", "quickstart", "main.go"), "package main")
			case "syntax":
				write(t, filepath.Join(root, "examples", "quickstart", "main.go"), "// snippet:quickstart\n???\n// endsnippet:quickstart\n")
			case "symbol":
				entries = []section{{"Values", []string{"Missing"}, "quickstart", []string{"quickstart"}, ""}}
			case "currency_list":
				write(t, filepath.Join(root, "currency_codes.go"), "package validation\n")
				entries = sections[:1]
			case "description":
				entries = []section{{"Values", []string{"String"}, "quickstart", []string{"quickstart"}, "Missing.Method"}}
			case "demonstration":
				entries = []section{{"Values", []string{"Match"}, "quickstart", []string{"quickstart"}, ""}}
			case "duplicate":
				entries = append(append([]section(nil), sections...), sections[0])
			case "marker":
				write(t, filepath.Join(root, "docs", "README.tmpl"), "missing markers")
			case "write":
				if err := os.Mkdir(filepath.Join(root, "README.md"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			if name == "write" {
				if err := generate(root); err == nil {
					t.Fatal("write error ignored")
				}
				return
			}
			if _, err := render(root, entries); err == nil {
				t.Fatal("malformed source accepted")
			}
			if name == "root" {
				if err := generate(root); err == nil {
					t.Fatal("generation error ignored")
				}
			}
		})
	}
}

// TestMainFailure verifies the CLI fails visibly without silently leaving stale output.
func TestMainFailure(t *testing.T) {
	t.Chdir(t.TempDir())
	defer func() {
		if recover() == nil {
			t.Fatal("main did not fail")
		}
	}()
	main()
}

// TestExtract checks marker ambiguity, reversed boundaries, and indentation preservation.
func TestExtract(t *testing.T) {
	for _, source := range []string{"", "// snippet:a\n// snippet:a\n// endsnippet:a\n", "// endsnippet:a\n// snippet:a\n", "// snippet:a\n \t\n// endsnippet:a\n"} {
		if _, err := extract(source, "a"); err == nil {
			t.Fatalf("bad boundaries accepted: %q", source)
		}
	}
	got, err := extract("// snippet:a\n\tif true {\n\n\t\tprintln(1)\n\t}\n// endsnippet:a\n", "a")
	if err != nil || got != "if true {\n\n\tprintln(1)\n}\n" {
		t.Fatalf("%q: %v", got, err)
	}
}

// TestCalls requires demonstrations in the actual snippet rather than elsewhere in the file.
func TestCalls(t *testing.T) {
	if _, err := calls("???"); err == nil {
		t.Fatal("invalid syntax accepted")
	}
	found, err := calls("validation.String(\"name\", \"Ada\").Required()")
	if err != nil || !found["String"] || !found["Required"] || found["Email"] {
		t.Fatalf("%v: %v", found, err)
	}
}

// TestReceiverDocumentation prevents same-named methods from replacing unrelated example prose.
func TestReceiverDocumentation(t *testing.T) {
	symbols, err := readSymbols("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"StringField.Required", "TimeField.Required", "Input.UnmarshalJSON", "uniqueRule.ValidateContext"} {
		if symbols[name] == "" {
			t.Errorf("missing receiver-qualified documentation: %s", name)
		}
	}
	output, err := render("../..", sections)
	if err != nil {
		t.Fatal(err)
	}
	_, after, _ := strings.Cut(output, "### Required · RequiredIf")
	description, _, _ := strings.Cut(after, "```go")
	if !strings.Contains(description, symbols["StringField.Required"]) || strings.Contains(description, "zero time") {
		t.Fatal(description)
	}
}

// TestCurrencySelectorList verifies deterministic grouping and complete projection of the pinned identifiers.
func TestCurrencySelectorList(t *testing.T) {
	sample := map[string]string{"CurrencyField.USD": "USD", "CurrencyField.UGX": "UGX", "CurrencyField.EUR": "EUR", "CurrencyField.Min": "Min", "StringField.GBP": "GBP"}
	output, err := currencySelectors(sample)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "All 3 currency selectors") || !strings.Contains(output, "| U | `.UGX()` · `.USD()` |") || strings.Contains(output, "GBP") || strings.Contains(output, "Min") {
		t.Fatal(output)
	}
	if strings.Index(output, "| E |") > strings.Index(output, "| U |") {
		t.Fatal("currency groups are not sorted")
	}
	if _, err := currencySelectors(nil); err == nil {
		t.Fatal("empty currency API accepted")
	}
	symbols, err := readSymbols("../..")
	if err != nil {
		t.Fatal(err)
	}
	output, err = currencySelectors(symbols)
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("../../testdata/currency-codes.txt")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(string(source), "\n") {
		code := strings.TrimSpace(line)
		if code == "" || strings.HasPrefix(code, "#") {
			continue
		}
		count++
		if strings.Count(output, "`."+code+"()`") != 1 {
			t.Errorf("selector %s missing or repeated", code)
		}
	}
	if !strings.Contains(output, fmt.Sprintf("All %d currency selectors", count)) {
		t.Fatal("currency count differs from pinned source")
	}
}
