package rule

import (
	"fmt"
	"strings"
)

func (tr *TrimmedRule) Or(rule string, data map[string]interface{}) TrimmedRuleExecutionResult {
	trimmedRule := tr.String()

	if strings.Contains(trimmedRule, " OR ") && !isOperatorProtected(trimmedRule, " OR ") {
		parts := strings.SplitN(trimmedRule, " OR ", 2)
		if len(parts) == 2 {
			leftRule, rightRule := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			leftPassed, leftDetails, leftErr := EvaluateRule(leftRule, data)
			if leftErr != nil {
				return TrimmedRuleExecutionResult{
					Executed: true,
					Passed:   false,
					Detail:   fmt.Sprintf("Erro LHS OR ('%s'): %s", leftRule, leftDetails),
					Err:      leftErr,
				}
			}
			if leftPassed {
				return TrimmedRuleExecutionResult{
					Executed: true,
					Passed:   true,
					Detail:   fmt.Sprintf("(%s) OR ('%s' não avaliada) -> true", leftDetails, rightRule),
					Err:      nil,
				}
			}
			rightPassed, rightDetails, rightErr := EvaluateRule(rightRule, data)
			if rightErr != nil {
				return TrimmedRuleExecutionResult{
					Executed: true,
					Passed:   false,
					Detail:   fmt.Sprintf("Erro RHS OR ('%s'): %s", rightRule, rightDetails),
					Err:      rightErr,
				}
			}
			return TrimmedRuleExecutionResult{
				Executed: true,
				Passed:   rightPassed,
				Detail:   fmt.Sprintf("(%s) OR (%s) -> %t", leftDetails, rightDetails, rightPassed),
				Err:      nil,
			}
		}
	}

	return TrimmedRuleExecutionResult{
		Executed: false,
	}
}
