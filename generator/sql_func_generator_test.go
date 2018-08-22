package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/packagesx"
)

func TestSqlFuncGenerator(t *testing.T) {
	cwd, _ := os.Getwd()
	pkg, _ := packagesx.Load(filepath.Join(cwd, "./__examples__/database"))

	g := NewSqlFuncGenerator(pkg)
	g.WithComments = true
	g.WithTableInterfaces = true
	g.StructName = "User"
	g.Database = "DBTest"

	g.Scan()
	g.Output(cwd)
}
