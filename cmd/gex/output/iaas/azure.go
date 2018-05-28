package iaas

import (
	"fmt"
	"os"
	"strings"

	"github.com/afritzler/garden-examiner/cmd/gex/util"
	"github.com/afritzler/garden-examiner/pkg"
)

func init() {
	RegisterIaasHandler(&azure{}, "azure")
}

type azure struct {
}

func (this *azure) Execute(shoot gube.Shoot, config map[string]string, args ...string) error {
	err := util.ExecCmd("az login --service-principal -u " + string(config["clientID"]) + " -p " + string(config["clientSecret"]) + " --tenant " + string(config["tenantID"]))
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	err = util.ExecCmd(strings.Join(args, " "))
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	return nil
}
