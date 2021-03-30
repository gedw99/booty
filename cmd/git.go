package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.amplifyedge.org/booty-v2/dep"
	"os"
)

// Simple wrapper for git
func GitWrapperCmd(gw dep.GitWrapper) *cobra.Command {
	gitCmd := &cobra.Command{
		Use: "gw",
		DisableFlagParsing: true,
		Short: "Gee Dubs (gw) is a simple wrapper for git, see -help for subcommands",
	}
	gitCmd.DisableFlagParsing = true
	gitCmd.Flags().SetInterspersed(true)
	tagCmd := &cobra.Command{Use: "tag <subcommand>", Short: "tag a release"}
	tagSubCmds := []*cobra.Command{
		{
			Use:     "new",
			Short:   "new <tag_name> <tag_msg>, note that tag name has to follow semver",
			Example: "new 0.0.1 new_release",
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.CreateTag(args[0], args[1])
			},
		},
		{
			Use:     "push",
			Short:   "push tags to the upstream",
			Example: "push",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.PushTag()
			},
		},
		{
			Use:     "del <tag_name>",
			Short:   "delete tags to the upstream",
			Example: "del master",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.DeleteTag(args[0])
			},
		},
		{
			Use:     "latest",
			Short:   "print latest tag name",
			Example: "latest",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				info, err := gw.RepoInfo(wd)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(os.Stdout, info.LastTag)
				return err
			},
		},
	}
	tagCmd.AddCommand(tagSubCmds...)
	gitSubCommands := []*cobra.Command{
		tagCmd,
		{
			Use:     "reg <directories>",
			Short:   "register repos to db",
			Example: "reg '/home/user/git/sys-share' '/home/user/git/sys'",
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.RegisterRepos(args...)
			},
		},
		{
			Use:     "ref",
			Short:   "ref information (sha commit)",
			Example: "ref",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				info, err := gw.RepoInfo(wd)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(os.Stdout, info.CurrentRef)
				return err
			},
		},
		{
			Use:     "current-branch",
			Short:   "current branch information",
			Example: "current-branch",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				info, err := gw.RepoInfo(wd)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(os.Stdout, info.CurrentBranch)
				return err
			},
		},
		{
			Use: "fset <UPSTREAM_OWNER>",
			Example: `
			First run booty gw reg <repositories>
			then do:
					fset amplify-edge
			`,
			Short: "Sets up upstream fork, username, and email for current git directory",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.SetupFork(args[0])
			},
		},
		{
			Use:     "fup",
			Example: "fup",
			Short:   "fup does a git fetch upstream and merge the commits from upstream/master to current git directory",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.CatchupFork()
			},
		},
		{
			Use:     "fup-all",
			Example: "fup-all",
			Short:   "fup-all does a git fetch upstream and merge the commits from upstream/master to all registered repositories",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.CatchupAll()
			},
		},
		{
			Use:     "add",
			Short:   "stage paths to git",
			Example: "add ./go.mod ./go.sum README.md",
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.Stage(args...)
			},
		},
		{
			Use:     "add-all",
			Short:   "stage all to git",
			Example: "add-all",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.StageAll()
			},
		},
		{
			Use:     "commit",
			Short:   "git commit -m ",
			Example: "commit <msg>",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.Commit(args[0])
			},
		},
		{
			Use:     "push",
			Short:   "git push",
			Example: "push",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return gw.Push()
			},
		},
		{
			Use:     "print",
			Short:   "print current directory's git information",
			Example: "print",
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				info, err := gw.RepoInfo(wd)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(os.Stdout, info.String())
				return err
			},
		},
	}
	gitCmd.AddCommand(gitSubCommands...)
	return gitCmd
}
