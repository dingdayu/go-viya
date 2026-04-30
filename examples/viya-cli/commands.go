package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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
	cmd.AddCommand(newDataCommand(ioStreams))
	cmd.AddCommand(newFilesCommand(ioStreams))
	cmd.AddCommand(newJobsCommand(ioStreams))
	cmd.AddCommand(newReportsCommand(ioStreams))
	cmd.AddCommand(newDashboardCommand(ioStreams))
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

func newDataCommand(ioStreams cliIO) *cobra.Command {
	opts := &dataOptions{cfg: cliConfig{Timeout: 5 * time.Minute}}
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Upload and promote CAS data",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)
	cmd.AddCommand(newDataUploadCSVCommand(ioStreams, opts))
	cmd.AddCommand(newDataPromoteCommand(ioStreams, opts))
	return cmd
}

func newDataUploadCSVCommand(ioStreams cliIO, opts *dataOptions) *cobra.Command {
	var table casTableFlags
	var file string
	cmd := &cobra.Command{
		Use:   "upload-csv",
		Short: "Upload CSV data into a CAS table",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := table.require(); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			if err := requireFlag("file", file); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			content, err := readInputFile(file, ioStreams.stdin)
			if err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.UploadCSVToCASTable(ctx, table.server, table.caslib, table.table, content)
			})
		},
	}
	table.addFlags(cmd.Flags())
	cmd.Flags().StringVar(&file, "file", "", "Path to a CSV file to upload, or - for stdin")
	return cmd
}

func newDataPromoteCommand(ioStreams cliIO, opts *dataOptions) *cobra.Command {
	var table casTableFlags
	cmd := &cobra.Command{
		Use:   "promote",
		Short: "Promote a CAS table to global scope",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := table.require(); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.PromoteCASTable(ctx, table.server, table.caslib, table.table)
			})
		},
	}
	table.addFlags(cmd.Flags())
	return cmd
}

func newFilesCommand(ioStreams cliIO) *cobra.Command {
	opts := &filesOptions{cfg: cliConfig{Timeout: 5 * time.Minute}}
	cmd := &cobra.Command{
		Use:   "files",
		Short: "List, upload, and download Viya Files Service files",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)
	cmd.AddCommand(newFilesListCommand(ioStreams, opts))
	cmd.AddCommand(newFilesUploadCommand(ioStreams, opts))
	cmd.AddCommand(newFilesDownloadCommand(ioStreams, opts))
	return cmd
}

func newFilesListCommand(ioStreams cliIO, opts *filesOptions) *cobra.Command {
	var limit int
	var filterName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Viya Files Service files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetFiles(ctx, viya.FileListOptions{Limit: limit, FilterName: filterName})
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum files to return")
	cmd.Flags().StringVar(&filterName, "filter-name", "", "Optional file name substring filter")
	return cmd
}

func newFilesUploadCommand(ioStreams cliIO, opts *filesOptions) *cobra.Command {
	var name string
	var file string
	var contentType string
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file to the Viya Files Service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("name", name); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			if err := requireFlag("file", file); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			content, err := readInputFile(file, ioStreams.stdin)
			if err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.UploadFile(ctx, name, contentType, content)
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Name for the uploaded file")
	cmd.Flags().StringVar(&file, "file", "", "Path to a file to upload, or - for stdin")
	cmd.Flags().StringVar(&contentType, "content-type", "text/plain", "MIME type for the uploaded file")
	return cmd
}

func newFilesDownloadCommand(ioStreams cliIO, opts *filesOptions) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a Viya Files Service file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("id", id); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.DownloadFile(ctx, id)
			})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "File ID to download")
	return cmd
}

func newJobsCommand(ioStreams cliIO) *cobra.Command {
	opts := &jobsOptions{cfg: cliConfig{Timeout: 5 * time.Minute}}
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Submit and inspect Job Execution service jobs",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)
	cmd.AddCommand(newJobsSubmitCommand(ioStreams, opts))
	cmd.AddCommand(newJobsListCommand(ioStreams, opts))
	cmd.AddCommand(newJobsStatusCommand(ioStreams, opts))
	cmd.AddCommand(newJobsCancelCommand(ioStreams, opts))
	cmd.AddCommand(newJobsLogCommand(ioStreams, opts))
	return cmd
}

func newJobsSubmitCommand(ioStreams cliIO, opts *jobsOptions) *cobra.Command {
	var code string
	var file string
	var name string
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit SAS code as a Job Execution job",
		RunE: func(cmd *cobra.Command, args []string) error {
			sasCode, err := readCode(code, file, ioStreams.stdin)
			if err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			if strings.TrimSpace(sasCode) == "" {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, fmt.Errorf("SAS code is required; pass --code, --file, or stdin via --file -"))
			}
			return runViyaCommandWithConfig(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client, cfg cliConfig) (any, error) {
				return client.SubmitJobExecutionCode(ctx, viya.SubmitJobExecutionCodeRequest{
					Name:        name,
					Code:        sasCode,
					ContextName: cfg.ContextName,
				})
			})
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "SAS code to submit")
	cmd.Flags().StringVar(&file, "file", "", "Path to a .sas file to submit, or - for stdin")
	cmd.Flags().StringVar(&name, "name", "", "Job name")
	cmd.Flags().StringVar(&opts.cfg.ContextName, "context-name", "", "Compute context name")
	return cmd
}

func newJobsListCommand(ioStreams cliIO, opts *jobsOptions) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Job Execution jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetJobExecutionJobs(ctx, viya.ListOptions{Limit: limit})
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum jobs to return")
	return cmd
}

func newJobsStatusCommand(ioStreams cliIO, opts *jobsOptions) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get Job Execution job status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("id", id); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetJobExecutionJob(ctx, id)
			})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Job ID")
	return cmd
}

func newJobsCancelCommand(ioStreams cliIO, opts *jobsOptions) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a Job Execution job",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("id", id); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				if err := client.CancelJobExecutionJob(ctx, id); err != nil {
					return nil, err
				}
				return fmt.Sprintf("Job %s cancelled.", id), nil
			})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Job ID")
	return cmd
}

func newJobsLogCommand(ioStreams cliIO, opts *jobsOptions) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Retrieve a Job Execution job log",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("id", id); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetJobExecutionJobLog(ctx, id)
			})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Job ID")
	return cmd
}

func newReportsCommand(ioStreams cliIO) *cobra.Command {
	opts := &reportsOptions{cfg: cliConfig{Timeout: 5 * time.Minute}}
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "List reports and request report image rendering",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)
	cmd.AddCommand(newReportsListCommand(ioStreams, opts))
	cmd.AddCommand(newReportsGetCommand(ioStreams, opts))
	cmd.AddCommand(newReportsImageCommand(ioStreams, opts))
	return cmd
}

func newReportsListCommand(ioStreams cliIO, opts *reportsOptions) *cobra.Command {
	var limit int
	var filterName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Visual Analytics reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetReports(ctx, viya.ReportListOptions{Limit: limit, FilterName: filterName})
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum reports to return")
	cmd.Flags().StringVar(&filterName, "filter-name", "", "Optional report name substring filter")
	return cmd
}

func newReportsGetCommand(ioStreams cliIO, opts *reportsOptions) *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get report metadata and definition",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("id", id); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.GetReport(ctx, id)
			})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Report ID")
	return cmd
}

func newReportsImageCommand(ioStreams cliIO, opts *reportsOptions) *cobra.Command {
	var id string
	var sectionIndex int
	var size string
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Request report section image rendering",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("id", id); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.CreateReportImageJob(ctx, viya.ReportImageJobRequest{
					ReportID:     id,
					SectionIndex: sectionIndex,
					Size:         size,
				})
			})
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Report ID")
	cmd.Flags().IntVar(&sectionIndex, "section-index", 0, "Report section index")
	cmd.Flags().StringVar(&size, "size", "800x600", "Rendered image size")
	return cmd
}

func newDashboardCommand(ioStreams cliIO) *cobra.Command {
	opts := &dashboardOptions{cfg: cliConfig{Timeout: 5 * time.Minute}}
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Create Visual Analytics dashboards from agent-friendly specs",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)
	cmd.AddCommand(newDashboardCreateCommand(ioStreams, opts))
	return cmd
}

func newDashboardCreateCommand(ioStreams cliIO, opts *dashboardOptions) *cobra.Command {
	var server string
	var caslib string
	var table string
	var name string
	var folderURI string
	var specFile string
	var resultNameConflict string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Visual Analytics dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlag("name", name); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			if err := requireFlag("folder-uri", folderURI); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			if err := requireFlag("spec", specFile); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			content, err := readInputFile(specFile, ioStreams.stdin)
			if err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			var spec viya.DashboardSpec
			if err := json.Unmarshal(content, &spec); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, fmt.Errorf("decode dashboard spec: %w", err))
			}
			return runViyaCommand(ioStreams, opts.cfg, func(ctx context.Context, client *viya.Client) (any, error) {
				return client.CreateDashboard(ctx, viya.CreateDashboardRequest{
					Name:               name,
					FolderURI:          folderURI,
					ResultNameConflict: resultNameConflict,
					ServerID:           server,
					CaslibName:         caslib,
					TableName:          table,
					Spec:               spec,
				})
			})
		},
	}
	cmd.Flags().StringVar(&server, "server", "", "CAS server name or ID")
	cmd.Flags().StringVar(&caslib, "caslib", "", "CAS library name")
	cmd.Flags().StringVar(&table, "table", "", "CAS table name")
	cmd.Flags().StringVar(&name, "name", "", "Dashboard report name")
	cmd.Flags().StringVar(&folderURI, "folder-uri", "", "Target folder URI")
	cmd.Flags().StringVar(&specFile, "spec", "", "Path to dashboard JSON spec, or - for stdin")
	cmd.Flags().StringVar(&resultNameConflict, "result-name-conflict", "rename", "Name conflict behavior: abort, rename, or replace")
	return cmd
}

func readInputFile(path string, stdin io.Reader) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
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
