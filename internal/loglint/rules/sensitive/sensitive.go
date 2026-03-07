package sensitive

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/extract"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/model"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/ruleset"
	"golang.org/x/tools/go/analysis"
)

type Matcher struct {
	keywords []string
	regexps  []*regexp.Regexp
}

type DataRule struct {
	matcher Matcher
}

func NewDataRule(m Matcher) DataRule {
	return DataRule{matcher: m}
}

func (r DataRule) Check(pass *analysis.Pass, msg model.LogMessage, ctx ruleset.Context) {
	if !ctx.Config.Rules.SensitiveData {
		return
	}
	if msg.IsConst {
		CheckStatic(pass, msg.Expr, msg.StaticText, r.matcher, ctx.Config.Autofix.SensitiveData)
		return
	}
	CheckDynamic(pass, msg.Expr, r.matcher)
}

func NewMatcher(keywords []string, regex []string) (Matcher, error) {
	m := Matcher{
		keywords: normalizeKeywords(keywords),
		regexps:  make([]*regexp.Regexp, 0, len(regex)),
	}

	for _, raw := range regex {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return Matcher{}, err
		}
		m.regexps = append(m.regexps, re)
	}
	return m, nil
}

func CheckStatic(pass *analysis.Pass, expr ast.Expr, message string, matcher Matcher, withSuggestedFix bool) {
	lower := strings.ToLower(message)
	if !containsKeyWithSeparator(lower, matcher) {
		return
	}

	diagnostic := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message may expose sensitive data",
	}
	if withSuggestedFix {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			diagnostic.SuggestedFixes = []analysis.SuggestedFix{
				{
					Message: "replace with neutral message",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     lit.Pos(),
							End:     lit.End(),
							NewText: []byte(strconv.Quote("sensitive data redacted")),
						},
					},
				},
			}
		}
	}
	pass.Report(diagnostic)
}

func CheckDynamic(pass *analysis.Pass, expr ast.Expr, matcher Matcher) {
	if containsReference(pass, expr, matcher) {
		pass.Reportf(expr.Pos(), "log message may expose sensitive data")
	}
}

func containsReference(pass *analysis.Pass, expr ast.Expr, matcher Matcher) bool {
	found := false

	ast.Inspect(expr, func(node ast.Node) bool {
		if node == nil || found {
			return false
		}

		switch n := node.(type) {
		case *ast.BasicLit:
			if n.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(n.Value)
			if err != nil {
				return true
			}
			if containsKeyWithSeparator(strings.ToLower(text), matcher) {
				found = true
				return false
			}
		case *ast.Ident:
			if isWord(n.Name, matcher) {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			if isWord(n.Sel.Name, matcher) {
				found = true
				return false
			}
		case *ast.CallExpr:
			if extract.LooksLikeSprintf(pass, n) {
				for _, arg := range n.Args {
					if containsReference(pass, arg, matcher) {
						found = true
						return false
					}
				}
			}
		}

		return true
	})

	return found
}

func containsKeyWithSeparator(s string, matcher Matcher) bool {
	for _, word := range matcher.keywords {
		if !strings.Contains(s, word) {
			continue
		}
		if strings.Contains(s, word+":") || strings.Contains(s, word+"=") {
			return true
		}
	}
	for _, re := range matcher.regexps {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func isWord(name string, matcher Matcher) bool {
	lower := strings.ToLower(name)
	lower = strings.ReplaceAll(lower, "-", "_")
	for _, word := range matcher.keywords {
		if lower == word || strings.Contains(lower, word) {
			return true
		}
	}
	for _, re := range matcher.regexps {
		if re.MatchString(lower) {
			return true
		}
	}
	return false
}

func normalizeKeywords(keywords []string) []string {
	out := make([]string, 0, len(keywords))
	for _, word := range keywords {
		w := strings.ToLower(strings.TrimSpace(word))
		if w == "" {
			continue
		}
		out = append(out, w)
	}
	return out
}
