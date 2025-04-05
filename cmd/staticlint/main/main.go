package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"

	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
)

const configFileName = `multichecker_config.json`

type Config struct {
	Staticcheck []string `json:"staticcheck"`
}

func main() {
	appfile, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), configFileName))
	if err != nil {
		log.Fatal(err)
	}
	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}

	res := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}

	// res = appendStaticcheckAnalyzers(res, cfg.Staticcheck)

	multichecker.Main(res...)
}

func appendStaticcheckAnalyzers(analyzers []*analysis.Analyzer, analyzerNames []string) []*analysis.Analyzer {
	checks := make(map[string]bool)
	for _, analyzerName := range analyzerNames {
		checks[analyzerName] = true
	}
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			analyzers = append(analyzers, v.Analyzer)
		}
	}
	return analyzers
}
