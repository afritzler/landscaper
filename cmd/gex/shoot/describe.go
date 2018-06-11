package shoot

import (
	"fmt"

	"github.com/mandelsoft/cmdint/pkg/cmdint"

	"github.com/afritzler/garden-examiner/cmd/gex/cmdline"
	"github.com/afritzler/garden-examiner/cmd/gex/context"
	"github.com/afritzler/garden-examiner/cmd/gex/iaas"
	"github.com/afritzler/garden-examiner/cmd/gex/output"
	"github.com/afritzler/garden-examiner/cmd/gex/util"
	"github.com/afritzler/garden-examiner/pkg"
)

func init() {
	filters.AddOptions(cmdline.AddAsVerb(GetCmdTab(), "describe", describe).CmdDescription(
		"describe shoot(s)",
	).
		CmdArgDescription("[<shoot>]"))
}

func describe(opts *cmdint.Options) error {
	return cmdline.ExecuteOutput(opts, NewDescribeOutput(), TypeHandler)
}

/////////////////////////////////////////////////////////////////////////////

type describe_output struct {
	*output.ElementOutput
}

func NewDescribeOutput() *describe_output {
	o := &describe_output{}
	o.ElementOutput = output.NewElementOutput(nil)
	return o
}

func (this *describe_output) Out(ctx *context.Context) error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		fmt.Printf("---\n")
		Describe(i.Next().(gube.Shoot))
	}
	return nil
}

func Describe(s gube.Shoot) error {
	attrs := util.NewAttributeSet()
	attrs.Attribute("Shoot", s.GetName().GetName())
	attrs.Attribute("Project", s.GetName().GetProjectName())
	attrs.Attributef("Profile", "%s (%s)", s.GetProfileName(), s.GetInfrastructure())
	attrs.Attribute("Seed Namespace", s.GetNamespaceInSeed())
	attrs.Attributef("API Server", "https://api.%s", s.GetDomainName())
	host, err := s.GetIngressHostFromSeed("alertmanager")
	if err == nil {
		attrs.Attribute("Alert Manager", "https://"+host)
	}
	host, err = s.GetIngressHostFromSeed("grafana")
	if err == nil {
		attrs.Attribute("Grafana", "https://"+host)
	}
	host, err = s.GetIngressHostFromSeed("prometheus")
	if err == nil {
		attrs.Attribute("Prometheus", "https://"+host)
	}
	user, pass, err := s.GetBasicAuth()
	if err != nil {
		attrs.Attributef("Basic Auth", "%s", err)
	} else {
		attrs.Attributef("Basic Auth", "%s (%s)", user, pass)
	}
	cnt := "unknown"
	c, err := s.GetNodeCount()
	if err == nil {
		cnt = fmt.Sprintf("%d", c)
	}
	attrs.Attribute("Number of Nodes", cnt)
	attrs.Attribute("State", s.GetState())
	cond := s.GetConditionErrors()
	if cond != nil {
		for c, m := range cond {
			attrs.Attribute("  "+c, m)
		}
	} else {
	}
	attrs.PrintAttributes()
	//iaas, err := s.GetIaaSInfo()
	iaas.Describe(s)

	e := s.GetError()
	if e != "" {
		fmt.Printf("Error: %s\n", s.GetError())
	}
	return nil
}
