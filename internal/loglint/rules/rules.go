package rules

import (
	"context"
	"go/ast"
	"go/token"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/config"
	"github.com/S1FFFkA/selectel-testcase-linter/internal/loglint/extract"
	go_translate "github.com/dinhcanh303/go_translate"
	"golang.org/x/tools/go/analysis"
)

type Context struct {
	Config     config.Config
	Matcher    SensitiveMatcher
	Translator go_translate.Translator
	UnifiedFix string
}

type Rule interface {
	Check(pass *analysis.Pass, msg extract.Message, ctx Context)
}

type LowercaseRule struct{}
type EnglishRule struct{}
type NoSpecialsRule struct{}
type SensitiveRule struct{}

func NewLowercaseRule() LowercaseRule { return LowercaseRule{} }
func NewEnglishRule() EnglishRule     { return EnglishRule{} }
func NewNoSpecialsRule() NoSpecialsRule {
	return NoSpecialsRule{}
}
func NewSensitiveRule() SensitiveRule { return SensitiveRule{} }

func (r LowercaseRule) Check(pass *analysis.Pass, msg extract.Message, ctx Context) {
	if !ctx.Config.Rules.LowercaseStart || msg.StaticText == "" {
		return
	}
	withFix := ctx.Config.Autofix.LowercaseStart && shouldAttachFixForLowercase(msg, ctx)
	checkStartsWithLowercase(pass, msg.Expr, msg.StaticText, withFix, ctx.UnifiedFix)
}

func (r EnglishRule) Check(pass *analysis.Pass, msg extract.Message, ctx Context) {
	if !ctx.Config.Rules.EnglishOnly || msg.StaticText == "" {
		return
	}
	withFix := ctx.Config.Autofix.EnglishOnly && shouldAttachFixForEnglish(msg, ctx)
	checkEnglishOnlyWithFix(pass, msg.Expr, msg.StaticText, withFix, ctx.Translator, ctx.UnifiedFix)
}

func (r NoSpecialsRule) Check(pass *analysis.Pass, msg extract.Message, ctx Context) {
	if !ctx.Config.Rules.NoSpecials || msg.StaticText == "" {
		return
	}
	withFix := ctx.Config.Autofix.NoSpecials && shouldAttachFixForNoSpecials(msg, ctx)
	checkNoSpecialsOrEmoji(pass, msg.Expr, msg.StaticText, msg.IsFormat, withFix, ctx.UnifiedFix)
}

func (r SensitiveRule) Check(pass *analysis.Pass, msg extract.Message, ctx Context) {
	if !ctx.Config.Rules.SensitiveData {
		return
	}
	if msg.IsConst {
		withFix := ctx.Config.Autofix.SensitiveData && shouldAttachFixForSensitive(msg, ctx)
		checkSensitiveStatic(pass, msg.Expr, msg.StaticText, ctx.Matcher, withFix, ctx.UnifiedFix)
		return
	}
	checkSensitiveDynamic(pass, msg.Expr, ctx.Matcher)
}

func BuildSensitiveMatcher(keywords []string, regexPatterns []string) (SensitiveMatcher, error) {
	m := SensitiveMatcher{
		keywords: normalizeKeywords(keywords),
		regexps:  make([]*regexp.Regexp, 0, len(regexPatterns)),
	}
	for _, raw := range regexPatterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return SensitiveMatcher{}, err
		}
		m.regexps = append(m.regexps, re)
	}
	return m, nil
}

type SensitiveMatcher struct {
	keywords []string
	regexps  []*regexp.Regexp
}

func checkStartsWithLowercase(pass *analysis.Pass, expr ast.Expr, message string, withSuggestedFix bool, unifiedFix string) {
	if startsWithLowercase(message) {
		return
	}
	r, _ := utf8.DecodeRuneInString(message)

	diagnostic := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message should start with a lowercase letter",
	}
	if withSuggestedFix && unicode.IsLetter(r) {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			replacement := string(unicode.ToLower(r)) + message[utf8.RuneLen(r):]
			if unifiedFix != "" {
				replacement = unifiedFix
			}
			diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
				Message: "lowercase the first letter",
				TextEdits: []analysis.TextEdit{{
					Pos:     lit.Pos(),
					End:     lit.End(),
					NewText: []byte(strconv.Quote(replacement)),
				}},
			}}
		}
	}
	pass.Report(diagnostic)
}

func checkEnglishOnlyWithFix(pass *analysis.Pass, expr ast.Expr, message string, withSuggestedFix bool, translator go_translate.Translator, unifiedFix string) {
	if isEnglishOnly(message) {
		return
	}
	diagnostic := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message should be in English only",
	}
	if withSuggestedFix && translator != nil {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			translated := translateToEnglish(message, translator)
			if unifiedFix != "" {
				translated = unifiedFix
			}
			if translated != "" && translated != message {
				diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
					Message: "translate to English",
					TextEdits: []analysis.TextEdit{{
						Pos:     lit.Pos(),
						End:     lit.End(),
						NewText: []byte(strconv.Quote(translated)),
					}},
				}}
			}
		}
	}
	pass.Report(diagnostic)
}

func checkNoSpecialsOrEmoji(pass *analysis.Pass, expr ast.Expr, message string, allowFormatVerb bool, withSuggestedFix bool, unifiedFix string) {
	diagnostic, found := findNoSpecialsOrEmojiViolation(expr, message, allowFormatVerb)
	if !found {
		return
	}
	if withSuggestedFix {
		applyNoSpecialsOrEmojiFix(&diagnostic, expr, message, allowFormatVerb, unifiedFix)
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

func applyNoSpecialsOrEmojiFix(diagnostic *analysis.Diagnostic, expr ast.Expr, message string, allowFormatVerb bool, unifiedFix string) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}
	sanitized := sanitizeMessage(message, allowFormatVerb)
	if unifiedFix != "" {
		sanitized = unifiedFix
	}
	if sanitized == message {
		return
	}
	diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
		Message: "remove special characters",
		TextEdits: []analysis.TextEdit{{
			Pos:     lit.Pos(),
			End:     lit.End(),
			NewText: []byte(strconv.Quote(sanitized)),
		}},
	}}
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
		if allowFormatVerb && r == '%' && i+1 < len(runes) {
			i++
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
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ':' || r == '=' || r == '_' {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if allowFormatVerb && r == '%' && i+1 < len(runes) {
			b.WriteRune(r)
			i++
			b.WriteRune(runes[i])
			lastSpace = false
			continue
		}
		if unicode.IsSpace(r) && !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

func checkSensitiveStatic(pass *analysis.Pass, expr ast.Expr, message string, matcher SensitiveMatcher, withSuggestedFix bool, unifiedFix string) {
	if !containsSensitiveKeyWithSeparator(strings.ToLower(message), matcher) {
		return
	}
	diagnostic := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message may expose sensitive data",
	}
	if withSuggestedFix {
		if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			replacement := "sensitive data redacted"
			if unifiedFix != "" {
				replacement = unifiedFix
			}
			diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
				Message: "replace with neutral message",
				TextEdits: []analysis.TextEdit{{
					Pos:     lit.Pos(),
					End:     lit.End(),
					NewText: []byte(strconv.Quote(replacement)),
				}},
			}}
		}
	}
	pass.Report(diagnostic)
}

func checkSensitiveDynamic(pass *analysis.Pass, expr ast.Expr, matcher SensitiveMatcher) {
	if containsSensitiveReference(pass, expr, matcher) {
		pass.Reportf(expr.Pos(), "log message may expose sensitive data")
	}
}

func containsSensitiveReference(pass *analysis.Pass, expr ast.Expr, matcher SensitiveMatcher) bool {
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
			if containsSensitiveKeyWithSeparator(strings.ToLower(text), matcher) {
				found = true
				return false
			}
		case *ast.Ident:
			if isSensitiveWord(n.Name, matcher) {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			if isSensitiveWord(n.Sel.Name, matcher) {
				found = true
				return false
			}
		case *ast.CallExpr:
			if extract.LooksLikeSprintf(pass, n) {
				for _, arg := range n.Args {
					if containsSensitiveReference(pass, arg, matcher) {
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

func containsSensitiveKeyWithSeparator(s string, matcher SensitiveMatcher) bool {
	for _, word := range matcher.keywords {
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

func isSensitiveWord(name string, matcher SensitiveMatcher) bool {
	lower := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
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
		if w != "" {
			out = append(out, w)
		}
	}
	return out
}

func translateToEnglish(text string, translator go_translate.Translator) string {
	result, err := translator.TranslateText(context.Background(), []string{text}, "en")
	if err != nil {
		log.Println("translate error:", err)
		return ""
	}
	if len(result) == 0 {
		return ""
	}
	return result[0]
}

func BuildUnifiedLiteralFix(msg extract.Message, ctx Context) string {
	lit, ok := msg.Expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING || msg.StaticText == "" {
		return ""
	}
	original := msg.StaticText
	fixed := original

	if ctx.Config.Autofix.EnglishOnly && ctx.Config.Rules.EnglishOnly && !isEnglishOnly(fixed) && ctx.Translator != nil {
		if translated := translateToEnglish(fixed, ctx.Translator); translated != "" {
			fixed = translated
		}
	}
	if ctx.Config.Autofix.SensitiveData && ctx.Config.Rules.SensitiveData && containsSensitiveKeyWithSeparator(strings.ToLower(fixed), ctx.Matcher) {
		fixed = "sensitive data redacted"
	}
	if ctx.Config.Autofix.NoSpecials && ctx.Config.Rules.NoSpecials && hasDisallowedRune(fixed, msg.IsFormat) {
		fixed = sanitizeMessage(fixed, msg.IsFormat)
	}
	if ctx.Config.Autofix.LowercaseStart && ctx.Config.Rules.LowercaseStart && !startsWithLowercase(fixed) {
		r, _ := utf8.DecodeRuneInString(fixed)
		if unicode.IsLetter(r) {
			fixed = string(unicode.ToLower(r)) + fixed[utf8.RuneLen(r):]
		}
	}
	if fixed == original {
		return ""
	}
	return fixed
}

func shouldAttachFixForLowercase(msg extract.Message, ctx Context) bool {
	if !ctx.Config.Rules.LowercaseStart {
		return false
	}
	return !startsWithLowercase(msg.StaticText)
}

func shouldAttachFixForEnglish(msg extract.Message, ctx Context) bool {
	if !ctx.Config.Rules.EnglishOnly {
		return false
	}
	if shouldAttachFixForLowercase(msg, ctx) {
		return false
	}
	return !isEnglishOnly(msg.StaticText)
}

func shouldAttachFixForNoSpecials(msg extract.Message, ctx Context) bool {
	if !ctx.Config.Rules.NoSpecials {
		return false
	}
	if shouldAttachFixForLowercase(msg, ctx) || shouldAttachFixForEnglish(msg, ctx) {
		return false
	}
	return hasDisallowedRune(msg.StaticText, msg.IsFormat)
}

func shouldAttachFixForSensitive(msg extract.Message, ctx Context) bool {
	if !ctx.Config.Rules.SensitiveData || !msg.IsConst {
		return false
	}
	if shouldAttachFixForLowercase(msg, ctx) || shouldAttachFixForEnglish(msg, ctx) || shouldAttachFixForNoSpecials(msg, ctx) {
		return false
	}
	return containsSensitiveKeyWithSeparator(strings.ToLower(msg.StaticText), ctx.Matcher)
}

func startsWithLowercase(message string) bool {
	r, _ := utf8.DecodeRuneInString(message)
	if r == utf8.RuneError || r == 0 {
		return true
	}
	return unicode.IsLower(r)
}

func isEnglishOnly(message string) bool {
	for _, r := range message {
		if unicode.IsLetter(r) && !unicode.In(r, unicode.Latin) {
			return false
		}
	}
	return true
}
