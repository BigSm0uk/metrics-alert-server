package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// StructInfo содержит информацию о структуре для генерации
type StructInfo struct {
	PackageName string
	StructName  string
	Fields      []FieldInfo
}

// FieldInfo содержит информацию о поле структуры
type FieldInfo struct {
	Name     string
	Type     string
	IsPtr    bool
	IsSlice  bool
	IsMap    bool
	ElemType string // для указателей, слайсов
	KeyType  string // для map
	ValType  string // для map
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Получаем корневую директорию проекта
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	// Сканируем все пакеты
	packages, err := scanPackages(rootDir)
	if err != nil {
		return fmt.Errorf("ошибка сканирования пакетов: %w", err)
	}

	// Для каждого пакета генерируем reset.gen.go
	for pkgPath, structs := range packages {
		if len(structs) == 0 {
			continue
		}

		if err := generateResetFile(pkgPath, structs); err != nil {
			return fmt.Errorf("ошибка генерации для %s: %w", pkgPath, err)
		}
		fmt.Printf("✅ Сгенерирован %s/reset.gen.go (%d структур)\n", pkgPath, len(structs))
	}

	return nil
}

// scanPackages сканирует все пакеты и находит структуры с комментарием // generate:reset
func scanPackages(rootDir string) (map[string][]StructInfo, error) {
	packages := make(map[string][]StructInfo)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем некоторые директории
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" ||
				name == "testdata" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		// Обрабатываем только .go файлы (кроме тестов и сгенерированных)
		if !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") ||
			strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		// Парсим файл
		structs, err := parseFile(path)
		if err != nil {
			return fmt.Errorf("ошибка парсинга %s: %w", path, err)
		}

		if len(structs) > 0 {
			pkgDir := filepath.Dir(path)
			packages[pkgDir] = append(packages[pkgDir], structs...)
		}

		return nil
	})

	return packages, err
}

// parseFile парсит Go файл и находит структуры с комментарием // generate:reset
func parseFile(filename string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structs []StructInfo

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// Проверяем комментарий
		if !hasGenerateResetComment(genDecl.Doc) {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Парсим поля структуры
			fields := parseFields(structType)

			structs = append(structs, StructInfo{
				PackageName: node.Name.Name,
				StructName:  typeSpec.Name.Name,
				Fields:      fields,
			})
		}
	}

	return structs, nil
}

// hasGenerateResetComment проверяет наличие комментария // generate:reset
func hasGenerateResetComment(commentGroup *ast.CommentGroup) bool {
	if commentGroup == nil {
		return false
	}

	for _, comment := range commentGroup.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		if strings.HasPrefix(text, "generate:reset") {
			return true
		}
	}

	return false
}

// parseFields парсит поля структуры
func parseFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			// Встроенное поле, пропускаем
			continue
		}

		for _, name := range field.Names {
			// Пропускаем неэкспортируемые поля
			if !ast.IsExported(name.Name) {
				continue
			}

			fieldInfo := FieldInfo{
				Name: name.Name,
				Type: exprToString(field.Type),
			}

			// Анализируем тип поля
			analyzeType(field.Type, &fieldInfo)

			fields = append(fields, fieldInfo)
		}
	}

	return fields
}

// analyzeType анализирует тип поля
func analyzeType(expr ast.Expr, field *FieldInfo) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		field.IsPtr = true
		field.ElemType = exprToString(t.X)
		analyzeType(t.X, field)
	case *ast.ArrayType:
		if t.Len == nil {
			field.IsSlice = true
		}
		field.ElemType = exprToString(t.Elt)
	case *ast.MapType:
		field.IsMap = true
		field.KeyType = exprToString(t.Key)
		field.ValType = exprToString(t.Value)
	}
}

// exprToString преобразует ast.Expr в строку
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

// generateResetFile генерирует файл reset.gen.go для пакета
func generateResetFile(pkgPath string, structs []StructInfo) error {
	var buf bytes.Buffer

	// Заголовок файла
	buf.WriteString("// Code generated by cmd/reset. DO NOT EDIT.\n\n")

	if len(structs) > 0 {
		buf.WriteString(fmt.Sprintf("package %s\n\n", structs[0].PackageName))
	}

	// Генерируем методы Reset для каждой структуры
	for i, s := range structs {
		if i > 0 {
			buf.WriteString("\n")
		}
		generateResetMethod(&buf, s)
	}

	// Форматируем код
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("ошибка форматирования: %w\n%s", err, buf.String())
	}

	// Записываем файл
	outputPath := filepath.Join(pkgPath, "reset.gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	return nil
}

// generateResetMethod генерирует метод Reset для структуры
func generateResetMethod(buf *bytes.Buffer, s StructInfo) {
	buf.WriteString(fmt.Sprintf("// Reset сбрасывает %s к начальным значениям\n", s.StructName))
	buf.WriteString(fmt.Sprintf("func (r *%s) Reset() {\n", s.StructName))
	buf.WriteString("\tif r == nil {\n")
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n")

	for _, field := range s.Fields {
		generateFieldReset(buf, field)
	}

	buf.WriteString("}\n")
}

// generateFieldReset генерирует код сброса для конкретного поля
func generateFieldReset(buf *bytes.Buffer, field FieldInfo) {
	fieldName := fmt.Sprintf("r.%s", field.Name)

	if field.IsPtr {
		// Обработка указателей
		buf.WriteString(fmt.Sprintf("\tif %s != nil {\n", fieldName))

		if field.IsSlice {
			// Указатель на слайс
			buf.WriteString(fmt.Sprintf("\t\t*%s = (*%s)[:0]\n", fieldName, fieldName))
		} else if field.IsMap {
			// Указатель на map
			buf.WriteString(fmt.Sprintf("\t\tclear(*%s)\n", fieldName))
		} else {
			// Указатель на примитив или структуру
			if isStruct(field.ElemType) {
				// Проверяем наличие метода Reset
				buf.WriteString(fmt.Sprintf("\t\tif resetter, ok := interface{}(*%s).(interface{ Reset() }); ok {\n", fieldName))
				buf.WriteString("\t\t\tresetter.Reset()\n")
				buf.WriteString("\t\t} else {\n")
				buf.WriteString(fmt.Sprintf("\t\t\t*%s = %s\n", fieldName, getZeroValue(field.ElemType)))
				buf.WriteString("\t\t}\n")
			} else {
				buf.WriteString(fmt.Sprintf("\t\t*%s = %s\n", fieldName, getZeroValue(field.ElemType)))
			}
		}

		buf.WriteString("\t}\n")
	} else if field.IsSlice {
		// Слайс (не указатель)
		buf.WriteString(fmt.Sprintf("\t%s = %s[:0]\n", fieldName, fieldName))
	} else if field.IsMap {
		// Map (не указатель)
		buf.WriteString(fmt.Sprintf("\tclear(%s)\n", fieldName))
	} else if isStruct(field.Type) {
		// Структура (не указатель)
		buf.WriteString(fmt.Sprintf("\tif resetter, ok := interface{}(%s).(interface{ Reset() }); ok {\n", fieldName))
		buf.WriteString("\t\tresetter.Reset()\n")
		buf.WriteString("\t} else {\n")
		buf.WriteString(fmt.Sprintf("\t\t%s = %s\n", fieldName, getZeroValue(field.Type)))
		buf.WriteString("\t}\n")
	} else {
		// Примитивный тип
		buf.WriteString(fmt.Sprintf("\t%s = %s\n", fieldName, getZeroValue(field.Type)))
	}
}

// isStruct проверяет, является ли тип структурой
func isStruct(typeName string) bool {
	// Примитивные типы
	primitives := map[string]bool{
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
		"bool": true, "string": true,
		"byte": true, "rune": true,
		"complex64": true, "complex128": true,
	}

	typeName = strings.TrimPrefix(typeName, "*")
	typeName = strings.TrimPrefix(typeName, "[]")

	return !primitives[typeName]
}

// getZeroValue возвращает нулевое значение для типа
func getZeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"byte", "rune":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	case "string":
		return `""`
	case "complex64", "complex128":
		return "0"
	default:
		// Для структур и других типов
		return typeName + "{}"
	}
}
