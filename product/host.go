package product

import (
	"runtime"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/op/host"
)

// NewHost takes a product config and creates a Product containing all of the host's ops.
func NewHost(logger hclog.Logger, cfg Config) *Product {
	return &Product{
		l:      logger.Named("product"),
		Name:   Host,
		Ops:    HostOps(cfg.OS),
		Config: cfg,
	}
}

// HostOps checks the operating system and passes it into the operations, returning a list of ops to run.
func HostOps(os string) []*op.Op {
	if os == "auto" {
		os = runtime.GOOS
	}
	return []*op.Op{
		host.NewOS(os),
		host.NewDisk(),
		host.NewInfo(),
		host.NewMemory(),
		host.NewProcess(),
		host.NewNetwork(),
		host.NewEtcHosts(),
		host.NewIPTables(),
		host.NewProcFile(os),
		host.NewFSTab(os),
	}
}
