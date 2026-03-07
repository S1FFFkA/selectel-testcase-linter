package style

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/model"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/ruleset"
	"golang.org/x/tools/go/analysis"
)

type LowercaseRule struct{}

func NewLowercaseRule() LowercaseRule {
	return LowercaseRule{}
}

func (r LowercaseRule) Check(pass *analysis.Pass, msg model.LogMessage, ctx ruleset.Context) {
	if !ctx.Config.Rules.LowercaseStart || msg.StaticText == "" {
		return
	}
	CheckStartsWithLowercase(pass, msg.Expr, msg.StaticText, ctx.Config.Autofix.LowercaseStart)
}

type EnglishRule struct{}

func NewEnglishRule() EnglishRule {
	return EnglishRule{}
}

func (r EnglishRule) Check(pass *analysis.Pass, msg model.LogMessage, ctx ruleset.Context) {
	if !ctx.Config.Rules.EnglishOnly || msg.StaticText == "" {
		return
	}
	CheckEnglishOnly(pass, msg.Expr, msg.StaticText)
}

type NoSpecialsRule struct{}

func NewNoSpecialsRule() NoSpecialsRule {
	return NoSpecialsRule{}
}

func (r NoSpecialsRule) Check(pass *analysis.Pass, msg model.LogMessage, ctx ruleset.Context) {
	if !ctx.Config.Rules.NoSpecials || msg.StaticText == "" {
		return
	}
	CheckNoSpecialsOrEmoji(pass, msg.Expr, msg.StaticText, msg.IsFormat, ctx.Config.Autofix.NoSpecials)
}

func CheckStartsWithLowercase(pass *analysis.Pass, expr ast.Expr, message string, withSuggestedFix bool) {
	r, _ := utf8.DecodeRuneInString(message)
	if r == utf8.RuneError || r == 0 {
		return
	}
	if unicode.IsLower(r) {
		return
	}

	diagnostic := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message should start with a lowercase letter",
	}

	if withSuggestedFix && unicode.IsLetter(r) {
		lowered := string(unicode.ToLower(r)) + message[utf8.RuneLen(r):]
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			diagnostic.SuggestedFixes = []analysis.SuggestedFix{
				{
					Message: "lowercase the first letter",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     lit.Pos(),
							End:     lit.End(),
							NewText: []byte(strconv.Quote(lowered)),
						},
					},
				},
			}
		}
	}

	pass.Report(diagnostic)
}

func CheckEnglishOnly(pass *analysis.Pass, expr ast.Expr, message string) {
	for _, r := range message {
		if !unicode.IsLetter(r) {
			continue
		}
		if isLatinLetter(r) {
			continue
		}
		pass.Reportf(expr.Pos(), "log message should be in English only")
		return
	}
}

func isLatinLetter(r rune) bool {
	return unicode.In(r, unicode.Latin)
}

func CheckNoSpecialsOrEmoji(pass *analysis.Pass, expr ast.Expr, message string, allowFormatVerb bool, withSuggestedFix bool) {
	diagnostic, found := findNoSpecialsOrEmojiViolation(expr, message, allowFormatVerb)
	if !found {
		return
	}
	if withSuggestedFix {
		applyNoSpecialsOrEmojiFix(&diagnostic, expr, message, allowFormatVerb)
	}
	pass.Report(diagnostic)
}

func findNoSpecialsOrEmojiViolation(expr ast.Expr, message string, allowFormatVerb bool) (analysis.Diagnostic, bool) {
	if !hasDisallowedRune(message, allowFormatVerb) {
		return analysis.Diagnostic{}, false
	}
	return analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message should not contain special characters or emoji",
	}, true
}

func applyNoSpecialsOrEmojiFix(diagnostic *analysis.Diagnostic, expr ast.Expr, message string, allowFormatVerb bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}
	sanitized := sanitizeMessage(message, allowFormatVerb)
	if sanitized == message {
		return
	}
	diagnostic.SuggestedFixes = []analysis.SuggestedFix{
		{
			Message: "remove special characters",
			TextEdits: []analysis.TextEdit{
				{
					Pos:     lit.Pos(),
					End:     lit.End(),
					NewText: []byte(strconv.Quote(sanitized)),
				},
			},
		},
	}
}

func hasDisallowedRune(message string, allowFormatVerb bool) bool {
	runes := []rune(message)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			continue
		}
		if r == ':' || r == '=' || r == '_' {
			continue
		}
		if allowFormatVerb && r == '%' {
			if i+1 < len(runes) {
				i++
			}
			continue
		}
		return true
	}
	return false
}

func sanitizeMessage(message string, allowFormatVerb bool) string {
	var b strings.Builder
	lastSpace := false

	runes := []rune(message)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if r == ':' || r == '=' || r == '_' {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if allowFormatVerb && r == '%' {
			if i+1 < len(runes) {
				b.WriteRune(r)
				i++
				b.WriteRune(runes[i])
				lastSpace = false
			}
			continue
		}
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
	}
	return strings.TrimSpace(b.String())
}
