package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-courier/packagesx"
)

func TestTagGen(t *testing.T) {
	cwd, _ := os.Getwd()
	pkg, _ := packagesx.Load(filepath.Join(cwd, "./__examples__/database"))

	g := NewTagGenerator(pkg)
	g.WithDefaults = true

	g.Scan("User", "User2")
	g.Output(cwd)
}
