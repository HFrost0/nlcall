package function

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"runtime"
	"strings"
)

type FuncInfo struct {
	Name       string               `json:"name"`
	Comments   string               `json:"comments"`
	SourceCode string               `json:"source_code"`
	Params     map[string]ParamInfo `json:"-"` // ignore by json
}

type ParamInfo struct {
	Name  string
	Index int
}

func GetFunctionDetails(fn any) (*FuncInfo, error) {
	// check f is func
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("fn is not a function")
	}

	// 使用 runtime 获取函数的文件和起始行
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if pc == nil {
		return nil, fmt.Errorf("unable to get function info")
	}

	file, startLine := pc.FileLine(pc.Entry())
	fmt.Printf("Function defined in file: %s at line: %d\n", file, startLine)

	// 打开源码文件
	srcData, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	src := string(srcData)

	// 解析源码文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source file: %w", err)
	}

	// 查找对应的函数定义
	var funcName, funcSource, funcComment string
	params := make(map[string]ParamInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			funcPos := fset.Position(fn.Pos())
			if funcPos.Line == startLine {
				// 获取函数名称
				funcName = fn.Name.Name

				// 提取函数源码
				start := fset.Position(fn.Pos()).Offset
				end := fset.Position(fn.End()).Offset
				funcSource = src[start:end]

				// 获取函数注释
				if fn.Doc != nil {
					var comments []string
					for _, comment := range fn.Doc.List {
						comments = append(comments, strings.TrimSpace(comment.Text))
					}
					funcComment = strings.Join(comments, "\n")
				}

				// 获取参数信息
				if fn.Type.Params != nil {
					for i, param := range fn.Type.Params.List {
						for _, name := range param.Names {
							params[name.Name] = ParamInfo{
								Name:  name.Name,
								Index: i,
							}
						}
					}
				}
				return false
			}
		}
		return true
	})

	if funcSource == "" {
		return nil, fmt.Errorf("function not found at line %d", startLine)
	}

	return &FuncInfo{
		Name:       funcName,
		Comments:   funcComment,
		SourceCode: funcSource,
		Params:     params,
	}, nil
}
