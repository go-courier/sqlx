package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/go-courier/codegen"
	"github.com/go-courier/codegen/formatx"
	"github.com/go-courier/packagesx"
	"github.com/sirupsen/logrus"
)

func NewTagGenerator(pkg *packagesx.Package) *TagGenerator {
	return &TagGenerator{
		pkg:   pkg,
		files: map[string]*ast.File{},
	}
}

type TagGenerator struct {
	WithDefaults bool
	pkg          *packagesx.Package
	files        map[string]*ast.File
}

func (g *TagGenerator) Scan(structNames ...string) {
	for ident, obj := range g.pkg.TypesInfo.Defs {
		if typeName, ok := obj.(*types.TypeName); ok {
			for _, structName := range structNames {
				if typeName.Name() == structName {
					if typeStruct, ok := typeName.Type().Underlying().(*types.Struct); ok {
						modifyTag(ident.Obj.Decl.(*ast.TypeSpec).Type.(*ast.StructType), typeStruct, g.WithDefaults)
						file := g.pkg.FileOf(ident)
						g.files[g.pkg.Fset.Position(file.Pos()).Filename] = file
					}
				}
			}
		}
	}
}

func (g *TagGenerator) writeFile(filename string, file *ast.File) {
	buf := bytes.NewBuffer(nil)

	if err := format.Node(buf, g.pkg.Fset, file); err != nil {
		panic(err)
	}

	data, err := formatx.Format(filename, buf.Bytes())
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		panic(err)
	}
}

func (g *TagGenerator) Output(cwd string) {
	for filename, file := range g.files {
		g.writeFile(filename, file)
	}
}

func modifyTag(structType *ast.StructType, typeStruct *types.Struct, withDefaults bool) {
	for i := 0; i < typeStruct.NumFields(); i++ {
		f := typeStruct.Field(i)
		if f.Anonymous() {
			continue
		}
		tags := getTags(typeStruct.Tag(i))
		astField := structType.Fields.List[i]

		if tags["db"] == "" {
			tags["db"] = fmt.Sprintf("F_%s", codegen.LowerSnakeCase(f.Name()))
		}
		if tags["json"] == "" {
			tags["json"] = codegen.LowerCamelCase(f.Name())
			switch f.Type().(type) {
			case *types.Basic:
				if f.Type().(*types.Basic).Kind() == types.Uint64 {
					tags["json"] = tags["json"] + ",string"
				}
			}
		}
		if tags["sql"] == "" {
			tpe := f.Type()
			switch deVendor(tpe.String()) {
			case "github.com/go-courier/sqlx/datatypes.MySQLDatetime":
				tags["sql"] = "datetime NOT NULL"
			case "github.com/go-courier/sqlx/datatypes.MySQLTimestamp":
				tags["sql"] = toSqlFromKind(types.Typ[types.Int64].Kind(), withDefaults)
			default:
				tpe, err := deref(tpe)
				if err != nil {
					logrus.Warnf("%s, make sure type of Field `%s` have sql.Valuer and sql.Scanner interface", err, f.Name())
				}
				switch tpe.(type) {
				case *types.Basic:
					tags["sql"] = toSqlFromKind(tpe.(*types.Basic).Kind(), withDefaults)
				default:
					tags["sql"] = WithDefaults("varchar(255) NOT NULL", withDefaults, "")
				}
			}
		}
		astField.Tag = &ast.BasicLit{Kind: token.STRING, Value: "`" + toTags(tags) + "`"}
	}
}

func deref(tpe types.Type) (types.Type, error) {
	switch tpe.(type) {
	case *types.Basic:
		return tpe.(*types.Basic), nil
	case *types.Struct, *types.Slice, *types.Array, *types.Map:
		return nil, fmt.Errorf("unsupport type %s", tpe)
	case *types.Pointer:
		return deref(tpe.(*types.Pointer).Elem())
	default:
		return deref(tpe.Underlying())
	}
}

func WithDefaults(dataType string, withDefaults bool, defaultValue string) string {
	if withDefaults {
		return dataType + fmt.Sprintf(" DEFAULT '%s'", defaultValue)
	}
	return dataType
}

func toSqlFromKind(kind types.BasicKind, withDefaults bool) string {
	switch kind {
	case types.Bool:
		return WithDefaults("tinyint(1) NOT NULL", withDefaults, "0")
	case types.Int8:
		return WithDefaults("tinyint NOT NULL", withDefaults, "0")
	case types.Int16:
		return WithDefaults("smallint NOT NULL", withDefaults, "0")
	case types.Int, types.Int32:
		return WithDefaults("int NOT NULL", withDefaults, "0")
	case types.Int64:
		return WithDefaults("bigint NOT NULL", withDefaults, "0")
	case types.Uint8:
		return WithDefaults("tinyint unsigned NOT NULL", withDefaults, "0")
	case types.Uint16:
		return WithDefaults("smallint unsigned NOT NULL", withDefaults, "0")
	case types.Uint, types.Uint32:
		return WithDefaults("int unsigned NOT NULL", withDefaults, "0")
	case types.Uint64:
		return WithDefaults("bigint unsigned NOT NULL", withDefaults, "0")
	case types.Float32:
		return WithDefaults("float NOT NULL", withDefaults, "0")
	case types.Float64:
		return WithDefaults("double NOT NULL", withDefaults, "0")
	default:
		// string
		return WithDefaults("varchar(255) NOT NULL", withDefaults, "")
	}
}

func toTags(tags map[string]string) (tag string) {
	names := make([]string, 0)
	for name := range tags {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		tag += fmt.Sprintf("%s:%s ", name, strconv.Quote(tags[name]))
	}
	return strings.TrimSpace(tag)
}

func getTags(tag string) (tags map[string]string) {
	tags = make(map[string]string)
	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}
		tags[name] = value

	}
	return
}
