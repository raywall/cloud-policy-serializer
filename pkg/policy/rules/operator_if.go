package rules

import (
	"fmt"
	"regexp"
	"strings"
)

func (tr *TrimmedRule) Condition(rule string, data map[string]interface{}) TrimmedRuleExecutionResult {
	trimmedRule := tr.String()

	if strings.HasPrefix(trimmedRule, "IF ") {
		parts := regexp.MustCompile(`^IF\s+(.+?)\s+THEN\s+(.+)$`).FindStringSubmatch(trimmedRule)
		if len(parts) != 3 {
			return TrimmedRuleExecutionResult{
				Executed: true,
				Passed:   false,
				Detail:   "",
				Err:      fmt.Errorf("regra IF...THEN inválida: %s", rule),
			}
		}
		conditionStr, actionStr := strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
		conditionMet, condDetails, errCond := EvaluateRule(conditionStr, data)
		if errCond != nil {
			return TrimmedRuleExecutionResult{
				Executed: true,
				Passed:   false,
				Detail:   fmt.Sprintf("Erro condição IF ('%s'): %s", conditionStr, condDetails),
				Err:      errCond,
			}
		}
		if conditionMet {
			actionPassed, actionDetails, actionErr := EvaluateRule(actionStr, data) // Ação pode ser SET com EXP
			if actionErr != nil {
				return TrimmedRuleExecutionResult{
					Executed: true,
					Passed:   false,
					Detail:   fmt.Sprintf("IF (%s) -> true, erro ação THEN ('%s'): %s", condDetails, actionStr, actionDetails),
					Err:      actionErr,
				}
			}
			return TrimmedRuleExecutionResult{
				Executed: true,
				Passed:   actionPassed,
				Detail:   fmt.Sprintf("IF (%s) -> true, THEN (%s) -> resultado ação: %t", condDetails, actionDetails, actionPassed),
				Err:      nil,
			}
		}
		return TrimmedRuleExecutionResult{
			Executed: true,
			Passed:   true,
			Detail:   fmt.Sprintf("IF (%s) -> false, ação ignorada: %s", condDetails, actionStr),
			Err:      nil,
		}
	}

	return TrimmedRuleExecutionResult{
		Executed: false,
	}
}
