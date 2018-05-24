package project

import (
	"github.com/mandelsoft/cmdint/pkg/cmdint"

	"github.com/afritzler/garden-examiner/cmd/gex/util"
	"github.com/afritzler/garden-examiner/cmd/gex/verb"
)

func init() {
	filters.AddOptions(verb.Add(GetCmdTab(), "select", cmd_select).
		CmdDescription("select project").CmdArgDescription("<project>"))

}

func cmd_select(opts *cmdint.Options) error {
	return util.ExecuteOutput(opts, verb.NewSelectOutput(), TypeHandler)
}
