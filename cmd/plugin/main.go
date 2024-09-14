package main

import (
	"github.com/mkaganm/outdatedcheck/pkg/outdatedcheck"
	"golang.org/x/tools/go/analysis"
)

var AnalyzerPlugin analyzerPlugin

type analyzerPlugin struct{}

func (a analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		outdatedcheck.Analyzer,
	}
}
