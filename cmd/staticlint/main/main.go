package main

import (
	"github.com/kisielk/errcheck/errcheck"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/stylecheck"

	"honnef.co/go/tools/simple"

	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"

	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"

	"github.com/karamaru-alpha/copyloopvar"
)

func main() {
	res := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		loopclosure.Analyzer,
		nilness.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		copyloopvar.NewAnalyzer(),
	}

	for _, v := range staticcheck.Analyzers {
		res = append(res, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		res = append(res, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		res = append(res, v.Analyzer)
	}
	for _, v := range quickfix.Analyzers {
		res = append(res, v.Analyzer)
	}

	multichecker.Main(res...)
}
