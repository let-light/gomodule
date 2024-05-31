package configcenter

import "github.com/let-light/gomodule"

type Feature interface {
	gomodule.IModule
	HelloWorld()
}
