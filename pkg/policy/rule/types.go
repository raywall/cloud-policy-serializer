package rule

import "strings"

// RuleExecutionResult armazena o resultado da execução de uma única regra.
type RuleExecutionResult struct {
	Rule    string
	Passed  bool
	Details string // Ex: "$.idade (20) >= 18 (true)" ou mensagem de erro
}

type TrimmedRule string

type TrimmedRuleExecutionResult struct {
	Executed bool
	Passed   bool
	Detail   string
	Err      error
}

func NewTrimmedRule(rule string) *TrimmedRule {
	trimmedRule := TrimmedRule(strings.TrimSpace(rule))
	return &trimmedRule
}

func (tr *TrimmedRule) String() string {
	return string(*tr)
}
