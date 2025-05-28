package rules

// RuleExecutionResult armazena o resultado da execução de uma única regra.
type RuleExecutionResultBackup struct {
	Rule    string
	Passed  bool
	Details string // Ex: "$.idade (20) >= 18 (true)" ou mensagem de erro
}
