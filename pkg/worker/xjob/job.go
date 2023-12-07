package job

import (
	"github.com/grey0904/jengine/pkg/flag"
)

func init() {
	flag.Register(
		&flag.StringFlag{
			Name:    "job",
			Usage:   "--job",
			Default: "",
		},
	)
}

// Runner ...
type Runner interface {
	Run()
}
