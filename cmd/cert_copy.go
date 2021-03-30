package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.amplifyedge.org/booty-v2/dep"
	"go.amplifyedge.org/booty-v2/internal/fileutil"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"path/filepath"
)

func CopyCertCommand(e dep.Executor) *cobra.Command {
	copyCertCmd := &cobra.Command{
		Use:   "cert-copy <destination_dir>",
		Short: "cert-copy <destination_dir>",
		Long:  "cert-copy copy mkcert certificates from our data directory to user specified destination dir",
		Args:  cobra.ExactArgs(1),
	}
	copyCertCmd.DisableFlagParsing = true
	copyCertCmd.Flags().SetInterspersed(true)
	copyCertCmd.RunE = func(cmd *cobra.Command, args []string) error {
		cname := "mkcert"
		c := e.Component(cname)
		if c == nil {
			return fmt.Errorf("%s is not a valid component", cname)
		}
		certDir := filepath.Join(osutil.GetDataDir(), cname)
		dst := args[0]
		_, err := fileutil.Copy(certDir, dst)
		return err
	}

	return copyCertCmd
}
