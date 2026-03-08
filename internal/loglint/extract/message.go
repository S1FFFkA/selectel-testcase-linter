package extract

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type Message struct {
	Expr       ast.Expr
	IsFormat   bool
	StaticText string
	IsConst    bool
}

func GetMessage(pass *analysis.Pass, call *ast.CallExpr) (Message, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Message{}, false
	}

	fnObj, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok {
		return Message{}, false
	}

	msgPos, isFormat, ok := messagePosition(pass, sel, fnObj)
	if !ok || msgPos >= len(call.Args) {
		return Message{}, false
	}

	msgExpr := call.Args[msgPos]
	if message, isConst := constString(pass, msgExpr); isConst {
		return Message{
			Expr:       msgExpr,
			IsFormat:   isFormat,
			StaticText: message,
			IsConst:    true,
		}, true
	}

	return Message{
		Expr:       msgExpr,
		IsFormat:   isFormat,
		StaticText: staticPrefixString(pass, msgExpr),
		IsConst:    false,
	}, true
}

func messagePosition(pass *analysis.Pass, sel *ast.SelectorExpr, fnObj *types.Func) (int, bool, bool) {
	method := sel.Sel.Name
	pkgPath := ""
	if fnObj.Pkg() != nil {
		pkgPath = fnObj.Pkg().Path()
	}

	sig, ok := fnObj.Type().(*types.Signature)
	if !ok {
		return 0, false, false
	}

	recv := sig.Recv()
	if recv == nil {
		if pkgPath != "log/slog" {
			return 0, false, false
		}
		switch method {
		case "Debug", "Info", "Warn", "Error":
			return 0, false, true
		case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
			return 1, false, true
		default:
			return 0, false, false
		}
	}

	recvPkg, recvName := receiverType(pass, sel)
	switch recvPkg {
	case "log/slog":
		if recvName != "Logger" {
			return 0, false, false
		}
		switch method {
		case "Debug", "Info", "Warn", "Error":
			return 0, false, true
		case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
			return 1, false, true
		default:
			return 0, false, false
		}
	case "go.uber.org/zap":
		switch recvName {
		case "Logger":
			switch method {
			case "Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal":
				return 0, false, true
			default:
				return 0, false, false
			}
		case "SugaredLogger":
			switch method {
			case "Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf":
				return 0, true, true
			case "Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw":
				return 0, false, true
			default:
				return 0, false, false
			}
		default:
			return 0, false, false
		}
	default:
		return 0, false, false
	}
}

func receiverType(pass *analysis.Pass, sel *ast.SelectorExpr) (string, string) {
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return "", ""
	}
	return namedTypeInfo(selection.Recv())
}

func namedTypeInfo(typ types.Type) (string, string) {
	for {
		ptr, ok := typ.(*types.Pointer)
		if !ok {
			break
		}
		typ = ptr.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return "", ""
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return "", obj.Name()
	}
	return obj.Pkg().Path(), obj.Name()
}

func constString(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

func staticPrefixString(pass *analysis.Pass, expr ast.Expr) string {
	if s, ok := constString(pass, expr); ok {
		return s
	}

	switch n := expr.(type) {
	case *ast.BinaryExpr:
		if n.Op != token.ADD {
			return ""
		}
		left := staticPrefixString(pass, n.X)
		if left == "" {
			return ""
		}
		right := staticPrefixString(pass, n.Y)
		return left + right
	case *ast.CallExpr:
		if !LooksLikeSprintf(pass, n) || len(n.Args) == 0 {
			return ""
		}
		format, _ := constString(pass, n.Args[0])
		return format
	default:
		return ""
	}
}

func LooksLikeSprintf(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	fnObj, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok || fnObj.Pkg() == nil {
		return false
	}
	return fnObj.Pkg().Path() == "fmt" && fnObj.Name() == "Sprintf"
}
