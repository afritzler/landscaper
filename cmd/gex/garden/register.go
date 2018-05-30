package garden

import (
	"fmt"
	"os"

	"github.com/afritzler/garden-examiner/cmd/gex/cmdline"
	"github.com/afritzler/garden-examiner/cmd/gex/context"
	"github.com/afritzler/garden-examiner/cmd/gex/util"
	"github.com/afritzler/garden-examiner/pkg"
	"github.com/mandelsoft/cmdint/pkg/cmdint"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	filters.AddOptions(cmdline.AddAsVerb(GetCmdTab(), "register", register).Raw().
		CmdDescription("run register for garden cluster").
		CmdArgDescription("[email]").
		ArgOption("garden"))
}

func register(opts *cmdint.Options) error {
	ctx := context.Get(opts)
	githubURL := ""
	if ctx.GardenSetConfig != nil {
		githubURL = ctx.GardenSetConfig.GetGithubURL()
	}
	switch len(opts.Arguments) {
	case 0:
		return register_garden(githubURL, ctx.Garden, "")
	case 1:
		return register_garden(githubURL, ctx.Garden, opts.Arguments[0])
	default:
		return fmt.Errorf("One optional email argument required")
	}
}

func getEmail(githubURL string) string {
	if githubURL == "" {
		return "null"
	}
	return util.ExecCmdReturnOutput("bash", "-c", "curl -ks "+githubURL+"/api/v3/users/"+os.Getenv("USER")+" | jq -r .email")
}

func register_garden(githubURL string, g gube.Garden, email string) error {
	if email == "" {
		email = getEmail(githubURL)
		if email == "null" {
			return fmt.Errorf("Could not read github email address")
		}
	}
	kubeset, err := g.GetClientset()
	if err != nil {
		return fmt.Errorf("failed to get garden clientset: %s", err)
	}
	clusterRoleBinding, err := kubeset.RbacV1().ClusterRoleBindings().Get("garden-administrators", metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, subject := range clusterRoleBinding.Subjects {
		if subject.Kind == "User" && subject.Name == email {
			return nil
		}
	}
	clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, rbacv1.Subject{Kind: "User", Name: email})
	_, err = kubeset.RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
	return err
}