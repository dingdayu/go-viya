package main

import (
	"context"
	"time"

	"github.com/dingdayu/go-viya"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newRootCommand(ioStreams cliIO) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "viya-cli",
		Short:         "Execute SAS code on SAS Viya for agent workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `viya-cli is a CLI example for agent frameworks.

It reads configuration from flags, environment variables, and
~/.sas/config.json plus ~/.sas/credentials.json.`,
	}
	cmd.SetOut(ioStreams.stdout)
	cmd.SetErr(ioStreams.stderr)
	cmd.AddCommand(newRunCommand(ioStreams))
	cmd.AddCommand(newCASCommand(ioStreams))
	return cmd
}

func newRunCommand(ioStreams cliIO) *cobra.Command {
	opts := &runOptions{
		cfg: cliConfig{
			Timeout:       5 * time.Minute,
			PollInterval:  2 * time.Second,
			IncludeOutput: true,
		},
	}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute SAS code in a Compute session",
		Example: `  viya-cli run --code "data _null_; put 'hello'; run;"
  viya-cli run --file program.sas
  viya-cli run --file - < program.sas`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSAS(ioStreams, opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.code, "code", "", "SAS code to execute")
	flags.StringVar(&opts.file, "file", "", "Path to a .sas file to execute, or - for stdin")
	addConfigFlags(flags, &opts.cfg)
	flags.StringVar(&opts.cfg.ContextID, "context-id", "", "Compute context ID")
	flags.StringVar(&opts.cfg.ContextName, "context-name", "", "Compute context name")
	flags.DurationVar(&opts.cfg.PollInterval, "poll-interval", opts.cfg.PollInterval, "Compute job polling interval")
	flags.BoolVar(&opts.cfg.KeepSession, "keep-session", opts.cfg.KeepSession, "Keep the Compute session after execution")
	flags.BoolVar(&opts.cfg.IncludeOutput, "include-output", opts.cfg.IncludeOutput, "Include log and listing text in JSON output")

	return cmd
}

func newCASCommand(ioStreams cliIO) *cobra.Command {
	opts := &casOptions{
		cfg: cliConfig{
			Timeout: 5 * time.Minute,
		},
	}

	cmd := &cobra.Command{
		Use:   "cas",
		Short: "Discover CAS servers, libraries, tables, columns, and rows",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)

	cmd.AddCommand(newCASServersCommand(ioStreams, opts))
	cmd.AddCommand(newCASLibsCommand(ioStreams, opts))
	cmd.AddCommand(newCASTablesCommand(ioStreams, opts))
	cmd.AddCommand(newCASTableInfoCommand(ioStreams, opts))
	cmd.AddCommand(newCASColumnsCommand(ioStreams, opts))
	cmd.AddCommand(newCASRowsCommand(ioStreams, opts))
	return cmd
}

func newCASServersCommand(ioStreams cliIO, opts *casOptions) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "servers",
		Short: "List CAS servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCASCommand(ioStreams, opts, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetCASServers(ctx, viya.ListOptions{Limit: limit})
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum servers to return")
	return cmd
}

func newCASLibsCommand(ioStreams cliIO, opts *casOptions) *cobra.Command {
	var server string
	var limit int
	cmd := &cobra.Command{
		Use:   "caslibs",
		Short: "List CAS libraries on a CAS server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("server", server); err != nil {
				return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runCASCommand(ioStreams, opts, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetCASLibs(ctx, server, viya.ListOptions{Limit: limit})
			})
		},
	}
	cmd.Flags().StringVar(&server, "server", "", "CAS server name or ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum caslibs to return")
	return cmd
}

func newCASTablesCommand(ioStreams cliIO, opts *casOptions) *cobra.Command {
	var server string
	var caslib string
	var limit int
	cmd := &cobra.Command{
		Use:   "tables",
		Short: "List tables in a CAS library",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("server", server); err != nil {
				return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			if err := requireFlag("caslib", caslib); err != nil {
				return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runCASCommand(ioStreams, opts, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetCASTables(ctx, server, caslib, viya.ListOptions{Limit: limit})
			})
		},
	}
	cmd.Flags().StringVar(&server, "server", "", "CAS server name or ID")
	cmd.Flags().StringVar(&caslib, "caslib", "", "CAS library name")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum tables to return")
	return cmd
}

func newCASTableInfoCommand(ioStreams cliIO, opts *casOptions) *cobra.Command {
	var table casTableFlags
	cmd := &cobra.Command{
		Use:   "table-info",
		Short: "Get metadata for a CAS table",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := table.require(); err != nil {
				return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runCASCommand(ioStreams, opts, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetCASTableInfo(ctx, table.server, table.caslib, table.table)
			})
		},
	}
	table.addFlags(cmd.Flags())
	return cmd
}

func newCASColumnsCommand(ioStreams cliIO, opts *casOptions) *cobra.Command {
	var table casTableFlags
	var limit int
	cmd := &cobra.Command{
		Use:   "columns",
		Short: "List columns for a CAS table",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := table.require(); err != nil {
				return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runCASCommand(ioStreams, opts, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetCASTableColumns(ctx, table.server, table.caslib, table.table, viya.ListOptions{Limit: limit})
			})
		},
	}
	table.addFlags(cmd.Flags())
	cmd.Flags().IntVar(&limit, "limit", 200, "Maximum columns to return")
	return cmd
}

func newCASRowsCommand(ioStreams cliIO, opts *casOptions) *cobra.Command {
	var table casTableFlags
	var start int
	var limit int
	cmd := &cobra.Command{
		Use:   "rows",
		Short: "Fetch sample rows from a CAS table",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := table.require(); err != nil {
				return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runCASCommand(ioStreams, opts, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetCASTableRows(ctx, table.server, table.caslib, table.table, viya.ListOptions{Start: start, Limit: limit})
			})
		},
	}
	table.addFlags(cmd.Flags())
	cmd.Flags().IntVar(&start, "start", 0, "Row offset")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum rows to return")
	return cmd
}

func addConfigFlags(flags *pflag.FlagSet, cfg *cliConfig) {
	flags.StringVar(&cfg.BaseURL, "base-url", "", "SAS Viya base URL")
	flags.StringVar(&cfg.ClientID, "client-id", "", "OAuth client ID")
	flags.StringVar(&cfg.ClientSecret, "client-secret", "", "OAuth client secret")
	flags.StringVar(&cfg.Username, "username", "", "SAS username for password flow")
	flags.StringVar(&cfg.Password, "password", "", "SAS password for password flow")
	flags.StringVar(&cfg.AccessToken, "access-token", "", "Bearer access token")
	flags.StringVar(&cfg.Profile, "profile", "", "Profile name in ~/.sas config files")
	flags.StringVar(&cfg.SASConfigDir, "sas-config-dir", "", "Directory containing config.json and credentials.json")
	flags.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "Maximum time to wait for SAS Viya API calls")
	flags.StringVarP(&cfg.Output, "output", "o", "text", "Output format: text or json")
}

type casTableFlags struct {
	server string
	caslib string
	table  string
}

func (f *casTableFlags) addFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.server, "server", "", "CAS server name or ID")
	flags.StringVar(&f.caslib, "caslib", "", "CAS library name")
	flags.StringVar(&f.table, "table", "", "CAS table name")
}

func (f casTableFlags) require() error {
	if err := requireFlag("server", f.server); err != nil {
		return err
	}
	if err := requireFlag("caslib", f.caslib); err != nil {
		return err
	}
	return requireFlag("table", f.table)
}
