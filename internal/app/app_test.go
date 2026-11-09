package app

import (
	"testing"

	"go.uber.org/fx"
)

func TestDependencyGraphs(t *testing.T) {
	for name, module := range map[string]fx.Option{"catalog": Catalog(), "reading": Reading(), "gateway": Gateway()} {
		t.Run(name, func(t *testing.T) {
			if err := fx.ValidateApp(module, fx.NopLogger); err != nil {
				t.Fatal(err)
			}
		})
	}
}
