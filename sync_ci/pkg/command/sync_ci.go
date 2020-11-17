package command

import (
	"context"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/subcommands"
	"github.com/pingcap/ci/sync_ci/pkg/model"
	"github.com/pingcap/ci/sync_ci/pkg/server"
)

type SyncCICommand struct {
	model.Config
}

func (*SyncCICommand) Name() string { return "sync-ci" }

func (*SyncCICommand) Synopsis() string { return "sync CI data to mysql" }

func (*SyncCICommand) Usage() string {
	return `sync-ci`
}

func (s *SyncCICommand) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.Dsn, "dsn", "root:123456@tcp(172.16.5.219:3306)/sync_ci_data", "dsn")
	f.StringVar(&s.Port, "port", "36000", "http service port")
}

func (s *SyncCICommand) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	server.NewServer(&s.Config).Run()
	return subcommands.ExitSuccess
}