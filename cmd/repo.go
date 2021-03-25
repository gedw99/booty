package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.amplifyedge.org/booty-v2/internal/repoclient"
	"go.amplifyedge.org/booty-v2/internal/reposerver"
	"net/http"
)

const (
	defaultPort       = 8085
	defaultServerAddr = "http://localhost:8085"
)

var (
	port       int
	serverAddr string
)

func PkgRepoServerCmd() *cobra.Command {

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start package repository server",
	}
	serveCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "default port to run the repository package server")
	serveCmd.RunE = func(cmd *cobra.Command, args []string) error {
		r := reposerver.NewServer()
		return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	}
	return serveCmd
}

func PkgRepoClientCmd() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Run upload/download client to the server.",
	}
	clientCmd.Flags().StringVarP(&serverAddr, "server", "s", defaultServerAddr, "server address")

	subcmds := []*cobra.Command{
		{
			Use:     "auth",
			Short:   "authenticate request",
			Example: "auth <user> <password>",
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return repoclient.AuthCli(serverAddr, args[0], args[1])
			},
		},
		{
			Use:     "ul",
			Short:   "upload file to repository",
			Example: "ul <location_to_file>",
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				name, err := repoclient.UploadCli(serverAddr, args[0])
				if err != nil {
					return err
				}
				fmt.Println("uploaded id: " + name)
				return nil
			},
		},
		{
			Use:     "dl",
			Short:   "download file from package repository",
			Example: "dl <file_id> <target_dir>",
			Args:    cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return repoclient.DownloadCli(serverAddr, args[0], args[1])
			},
		},
	}

	clientCmd.AddCommand(subcmds...)
	return clientCmd
}
