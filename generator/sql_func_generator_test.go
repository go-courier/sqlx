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

	for _, name := range []string{"User", "Org"} {
		g := NewSqlFuncGenerator(pkg)
		g.WithComments = true
		g.WithTableInterfaces = true
		g.WithMethods = true
		g.Database = "DBTest"
		g.StructName = name

		g.Scan()
		g.Output(cwd)
	}
}
