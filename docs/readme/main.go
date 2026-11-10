// Package main projects runnable workflows and GoDoc into the library README.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// section associates an application-facing API with its authoritative program snippets.
type section struct {
	group    string
	apis     []string
	program  string
	snippets []string
}

// sections curates consumer operations rather than enumerating implementation plumbing.
var sections = []section{
	{"Values", []string{"String", "Number", "Bool", "Validate"}, "quickstart", []string{"quickstart"}},
	{"Formats", []string{"IsEmail", "IsURL", "IsUUID", "IsUUIDv4", "IsUUIDv7", "IsIP", "IsIPv4", "IsIPv6", "IsJSON", "IsBase64"}, "formats", []string{"formats"}},
	{"Conditions", []string{"When"}, "patch", []string{"patch"}},
	{"Collections", []string{"Group", "Slice", "Each"}, "nested", []string{"nested"}},
	{"Custom rules", []string{"Using", "Invalid"}, "custom", []string{"custom_type", "custom"}},
	{"Context", []string{"ValidateContext", "Check"}, "context", []string{"context"}},
	{"Domain values", []string{"Value", "RuleFunc", "ContextRuleFunc"}, "domain", []string{"domain"}},
	{"Errors", []string{"Label", "Message", "First", "Messages", "Issues"}, "errors", []string{"errors"}},
	{"Constraints", []string{"Required", "RequiredIf", "NotBlank", "Min", "Max", "In", "Equal", "Match", "Accepted"}, "constraints", []string{"constraints"}},
	{"Text constraints", []string{"Len", "Contains", "NotContains", "StartsWith", "EndsWith", "NotEqual", "NotIn"}, "advanced", []string{"text"}},
	{"Numeric constraints", []string{"GreaterThan", "LessThan", "MultipleOf"}, "advanced", []string{"numbers"}},
	{"Collection constraints", []string{"Map", "Unique", "UniqueBy", "HasKey", "OnlyKeys"}, "advanced", []string{"collection_rules"}},
}

// main fails visibly rather than leaving stale generated documentation on disk.
func main() {
	if err := generate(".."); err != nil {
		panic(err)
	}
}

// generate writes only after all source and navigation checks have passed.
func generate(root string) error {
	output, err := render(root, sections)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "README.md"), []byte(output), 0644)
}

// render builds README sections from checked GoDoc symbols and executable sources.
func render(root string, entries []section) (string, error) {
	symbols, err := readSymbols(root)
	if err != nil {
		return "", err
	}
	template, err := os.ReadFile(filepath.Join(root, "docs", "README.tmpl"))
	if err != nil {
		return "", err
	}
	var index, examples, sources strings.Builder
	index.WriteString("| Group | Functions |\n| ---: | :--- |\n")
	lastGroup := ""
	seen := make(map[string]bool)
	programs := make(map[string]bool)
	quickstart := ""
	for _, entry := range entries {
		path := "examples/" + entry.program + "/main.go"
		source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return "", err
		}
		var code strings.Builder
		demonstrated := make(map[string]bool)
		for _, id := range entry.snippets {
			snippet, err := extract(string(source), id)
			if err != nil {
				return "", fmt.Errorf("%s: %w", path, err)
			}
			names, err := calls(snippet)
			if err != nil {
				return "", err
			}
			for name := range names {
				demonstrated[name] = true
			}
			code.WriteString(snippet + "\n")
		}
		if lastGroup != entry.group {
			fmt.Fprintf(&examples, "## %s\n\n", entry.group)
			lastGroup = entry.group
		}
		var links []string
		for _, api := range entry.apis {
			if symbols[api] == "" {
				return "", fmt.Errorf("missing GoDoc API %s", api)
			}
			if !demonstrated[api] {
				return "", fmt.Errorf("snippet does not demonstrate %s", api)
			}
			anchor := strings.ToLower(strings.ReplaceAll(api, ".", "-"))
			if seen[anchor] {
				return "", fmt.Errorf("duplicate API anchor %s", anchor)
			}
			seen[anchor] = true
			links = append(links, "["+api+"](#"+anchor+")")
			fmt.Fprintf(&examples, "<a id=\"%s\"></a>\n\n", anchor)
		}
		fmt.Fprintf(&index, "| **%s** | %s |\n", entry.group, strings.Join(links, " · "))
		fmt.Fprintf(&examples, "### %s\n\n%s\n\n```go\n%s```\n\n", strings.Join(entry.apis, " · "), strings.TrimSpace(symbols[entry.apis[0]]), code.String())
		if !programs[entry.program] {
			fmt.Fprintf(&sources, "- [%s](%s)\n", entry.program, path)
			programs[entry.program] = true
		}
		if entry.program == "quickstart" {
			quickstart = code.String()
		}
	}
	replacements := []struct{ key, value string }{
		{"{{QUICKSTART}}", quickstart}, {"{{INDEX}}", index.String()},
		{"{{EXAMPLES}}", examples.String()}, {"{{SOURCES}}", sources.String()},
	}
	output := string(template)
	for _, replacement := range replacements {
		if strings.Count(output, replacement.key) != 1 {
			return "", fmt.Errorf("template requires one %s marker", replacement.key)
		}
		output = strings.Replace(output, replacement.key, replacement.value, 1)
	}
	return output, nil
}

// readSymbols keeps public descriptions owned by GoDoc in their implementation files.
func readSymbols(root string) (map[string]string, error) {
	result := make(map[string]string)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, entry.Name()), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			switch d := declaration.(type) {
			case *ast.FuncDecl:
				if d.Doc != nil {
					result[d.Name.Name] = d.Doc.Text()
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if named, ok := spec.(*ast.TypeSpec); ok && d.Doc != nil {
						result[named.Name.Name] = d.Doc.Text()
					}
				}
			}
		}
	}
	return result, nil
}

// extract requires unique ordered snippet boundaries and removes only common indentation.
func extract(source, id string) (string, error) {
	begin, end := "// snippet:"+id+"\n", "// endsnippet:"+id+"\n"
	if strings.Count(source, begin) != 1 || strings.Count(source, end) != 1 {
		return "", fmt.Errorf("missing or duplicate snippet %s", id)
	}
	_, after, _ := strings.Cut(source, begin)
	code, _, ok := strings.Cut(after, end)
	if !ok {
		return "", fmt.Errorf("reversed snippet boundaries %s", id)
	}
	lines := strings.Split(strings.TrimRight(code, "\t \n"), "\n")
	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count := len(line) - len(strings.TrimLeft(line, "\t "))
		if indent < 0 || count < indent {
			indent = count
		}
	}
	if indent < 0 {
		return "", fmt.Errorf("empty snippet %s", id)
	}
	for i, line := range lines {
		if len(line) >= indent {
			lines[i] = line[indent:]
		}
	}
	return strings.Join(lines, "\n") + "\n", nil
}

// calls checks demonstrated syntax inside the snippet's own section.
func calls(code string) (map[string]bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "snippet.go", "package snippet\n"+code, 0)
	if err != nil {
		file, err = parser.ParseFile(token.NewFileSet(), "snippet.go", "package snippet\nfunc example(){\n"+code+"\n}", 0)
	}
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			function := call.Fun
			if index, ok := function.(*ast.IndexExpr); ok {
				function = index.X
			}
			if selector, ok := function.(*ast.SelectorExpr); ok {
				result[selector.Sel.Name] = true
			}
		}
		return true
	})
	return result, nil
}
