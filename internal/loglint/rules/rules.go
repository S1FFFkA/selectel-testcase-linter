package rules

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"
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
	// For any dynamic message (not pure constant), ':' and '=' are allowed.
	allowKVSeparators := !msg.IsConst
	allowFormatVerb := msg.IsFormat || isSprintfExpr(pass, msg.Expr)
	withFix := ctx.Config.Autofix.NoSpecials && shouldAttachFixForNoSpecials(pass, msg, ctx)
	checkNoSpecialsOrEmoji(pass, msg.Expr, msg.StaticText, allowFormatVerb, allowKVSeparators, withFix, ctx.UnifiedFix)
}

func (r SensitiveRule) Check(pass *analysis.Pass, msg extract.Message, ctx Context) {
	if !ctx.Config.Rules.SensitiveData {
		return
	}
	if !isSensitiveDynamicMessage(pass, msg, ctx.Matcher) {
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

	diagnostic := analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message should start with a lowercase letter",
	}
	if withSuggestedFix {
		lit, literalText, ok := fixTargetStringLiteral(pass, expr)
		if ok {
			first, _ := utf8.DecodeRuneInString(literalText)
			if unicode.IsLetter(first) {
				replacement := string(unicode.ToLower(first)) + literalText[utf8.RuneLen(first):]
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
		lit, literalText, ok := fixTargetStringLiteral(pass, expr)
		if ok {
			translated := translateToEnglish(literalText, translator)
			if unifiedFix != "" {
				translated = unifiedFix
			}
			if translated != "" && translated != literalText {
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

func checkNoSpecialsOrEmoji(pass *analysis.Pass, expr ast.Expr, message string, allowFormatVerb bool, allowKVSeparators bool, withSuggestedFix bool, unifiedFix string) {
	diagnostic, found := findNoSpecialsOrEmojiViolation(expr, message, allowFormatVerb, allowKVSeparators)
	if !found {
		return
	}
	if withSuggestedFix {
		applyNoSpecialsOrEmojiFix(&diagnostic, expr, message, allowFormatVerb, allowKVSeparators, unifiedFix)
	}
	pass.Report(diagnostic)
}

func findNoSpecialsOrEmojiViolation(expr ast.Expr, message string, allowFormatVerb bool, allowKVSeparators bool) (analysis.Diagnostic, bool) {
	if !hasDisallowedRune(message, allowFormatVerb, allowKVSeparators) {
		return analysis.Diagnostic{}, false
	}
	return analysis.Diagnostic{
		Pos:     expr.Pos(),
		End:     expr.End(),
		Message: "log message should not contain special characters or emoji",
	}, true
}

func applyNoSpecialsOrEmojiFix(diagnostic *analysis.Diagnostic, expr ast.Expr, message string, allowFormatVerb bool, allowKVSeparators bool, unifiedFix string) {
	lit, literalText, ok := fixTargetStringLiteral(nil, expr)
	if !ok {
		return
	}
	sanitized := sanitizeMessage(literalText, allowFormatVerb, allowKVSeparators)
	if unifiedFix != "" {
		sanitized = unifiedFix
	}
	if sanitized == literalText {
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

func hasDisallowedRune(message string, allowFormatVerb bool, allowKVSeparators bool) bool {
	runes := []rune(message)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			continue
		}
		if allowKVSeparators && (r == ':' || r == '=') {
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

func sanitizeMessage(message string, allowFormatVerb bool, allowKVSeparators bool) string {
	var b strings.Builder
	lastSpace := false
	runes := []rune(message)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) || (allowKVSeparators && (r == ':' || r == '=')) {
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
	canonicalName := canonicalSensitiveToken(lower)
	for _, word := range matcher.keywords {
		if canonicalName == canonicalSensitiveToken(word) {
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

func canonicalSensitiveToken(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
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

func BuildUnifiedLiteralFix(pass *analysis.Pass, msg extract.Message, ctx Context) string {
	_, original, ok := fixTargetStringLiteral(pass, msg.Expr)
	if !ok || original == "" {
		return ""
	}
	fixed := original

	if ctx.Config.Autofix.EnglishOnly && ctx.Config.Rules.EnglishOnly && !isEnglishOnly(fixed) && ctx.Translator != nil {
		if translated := translateToEnglish(fixed, ctx.Translator); translated != "" {
			fixed = translated
		}
	}
	allowKVSeparators := !msg.IsConst
	allowFormatVerb := msg.IsFormat || isSprintfExpr(pass, msg.Expr)
	if ctx.Config.Autofix.NoSpecials && ctx.Config.Rules.NoSpecials && hasDisallowedRune(fixed, allowFormatVerb, allowKVSeparators) {
		fixed = sanitizeMessage(fixed, allowFormatVerb, allowKVSeparators)
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

func shouldAttachFixForNoSpecials(pass *analysis.Pass, msg extract.Message, ctx Context) bool {
	if !ctx.Config.Rules.NoSpecials {
		return false
	}
	if shouldAttachFixForLowercase(msg, ctx) || shouldAttachFixForEnglish(msg, ctx) {
		return false
	}
	allowKVSeparators := !msg.IsConst
	allowFormatVerb := msg.IsFormat || isSprintfExpr(pass, msg.Expr)
	return hasDisallowedRune(msg.StaticText, allowFormatVerb, allowKVSeparators)
}

func fixTargetStringLiteral(pass *analysis.Pass, expr ast.Expr) (*ast.BasicLit, string, bool) {
	switch n := expr.(type) {
	case *ast.BasicLit:
		if n.Kind != token.STRING {
			return nil, "", false
		}
		value, err := strconv.Unquote(n.Value)
		if err != nil {
			return nil, "", false
		}
		return n, value, true
	case *ast.BinaryExpr:
		if n.Op != token.ADD {
			return nil, "", false
		}
		if lit, value, ok := fixTargetStringLiteral(pass, n.X); ok {
			return lit, value, true
		}
		return fixTargetStringLiteral(pass, n.Y)
	case *ast.CallExpr:
		if pass != nil && extract.LooksLikeSprintf(pass, n) && len(n.Args) > 0 {
			return fixTargetStringLiteral(pass, n.Args[0])
		}
		return nil, "", false
	default:
		return nil, "", false
	}
}

func isSensitiveDynamicMessage(pass *analysis.Pass, msg extract.Message, matcher SensitiveMatcher) bool {
	if msg.IsConst {
		return false
	}
	return hasVariableReference(pass, msg.Expr) && containsSensitiveReference(pass, msg.Expr, matcher)
}

func hasVariableReference(pass *analysis.Pass, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if node == nil || found {
			return false
		}
		switch n := node.(type) {
		case *ast.Ident:
			if obj, ok := pass.TypesInfo.Uses[n]; ok {
				if _, ok := obj.(*types.Var); ok {
					found = true
					return false
				}
			}
		case *ast.SelectorExpr:
			if sel := pass.TypesInfo.Selections[n]; sel != nil && sel.Kind() == types.FieldVal {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func isSprintfExpr(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	return extract.LooksLikeSprintf(pass, call)
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
