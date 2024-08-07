package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	goimports "github.com/incu6us/goimports-reviser/v3/reviser"
	"golang.org/x/mod/modfile"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"text/template"

	. "github.com/knaka/go-utils"
)

const generatorName = "gen-from-mod"

//goland:noinspection GoUnusedExportedType, GoUnnecessarilyExportedIdentifiers
type Module struct {
	Path string
}

//goland:noinspection GoUnusedExportedType, GoUnnecessarilyExportedIdentifiers
type Data struct {
	Module Module
}

func main() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(thisFilePath))

	goModBytes := V(os.ReadFile(filepath.Join(rootDir, "go.mod")))
	modulePath := modfile.ModulePath(goModBytes)
	data := Data{
		Module: Module{
			Path: modulePath,
		},
	}

	// foo.gen_bar.go.tmpl -> foo.gen_bar.go, foo.gen_bar.sql.tmpl -> foo.gen_bar.sql
	reTmplFileExt := regexp.MustCompile(`\.(` + generatorName + `)\.([a-zA-Z0-9_]+)\.tmpl$`)
	var tmplFilePaths []string
	V0(filepath.Walk(rootDir, func(path string, stat os.FileInfo, _ error) error {
		if stat.IsDir() {
			return nil
		}
		if reTmplFileExt.MatchString(path) {
			tmplFilePaths = append(tmplFilePaths, path)
		}
		return nil
	}))
	funcMap := sprig.HermeticTxtFuncMap()
	noticeFunc := func() string {
		return fmt.Sprintf("Code generated by %s. DO NOT EDIT.", generatorName)
	}
	funcMap["generatedContentNotice"] = noticeFunc
	funcMap["generatedCodeNotice"] = noticeFunc
	for _, tmplFilePath := range tmplFilePaths {
		tmplBody := string(V(os.ReadFile(tmplFilePath)))
		tmplTree := template.Must(template.New("").Funcs(funcMap).Parse(tmplBody))
		buf := bytes.Buffer{}
		V0(tmplTree.Execute(&buf, data))
		outFilePath := reTmplFileExt.ReplaceAllString(tmplFilePath, `.${1}.${2}`)
		V0(os.WriteFile(outFilePath, buf.Bytes(), 0o644))
		// Format Go source files.
		if filepath.Ext(outFilePath) == ".go" {
			sourceFile := goimports.NewSourceFile( /* data.PackagePath */ "", outFilePath)
			V0(goimports.WithRemovingUnusedImports(sourceFile))
			V0(goimports.WithCodeFormatting(sourceFile))
			fixedText, _, differs, err := sourceFile.Fix()
			if err != nil {
				// Rename the invalid file not to be used as a source file.
				V0(os.Rename(outFilePath, outFilePath+".err"))
				panic(err)
			}
			if differs {
				V0(os.WriteFile(outFilePath, fixedText, 0644))
			}
			Ignore(os.Remove(outFilePath + ".err"))
		}
	}
}
