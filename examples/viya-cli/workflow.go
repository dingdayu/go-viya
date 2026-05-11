package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dingdayu/go-viya"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const defaultWorkflowMaxParallel = 2

type workflowOptions struct {
	cfg            cliConfig
	userConfigFile string
	maxParallel    int
}

type workflowFlagOverrides struct {
	includeOutput bool
	keepSession   bool
	maxParallel   bool
	contextID     bool
	contextName   bool
}

type workflowUserConfig struct {
	ContextID   string            `json:"contextId,omitempty" yaml:"contextId,omitempty"`
	ContextName string            `json:"contextName,omitempty" yaml:"contextName,omitempty"`
	Autoexec    string            `json:"autoexec,omitempty" yaml:"autoexec,omitempty"`
	PreCode     string            `json:"preCode,omitempty" yaml:"preCode,omitempty"`
	PostCode    string            `json:"postCode,omitempty" yaml:"postCode,omitempty"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type workflowProjectConfig struct {
	Version     int                     `json:"version,omitempty" yaml:"version,omitempty"`
	Name        string                  `json:"name,omitempty" yaml:"name,omitempty"`
	MaxParallel int                     `json:"maxParallel,omitempty" yaml:"maxParallel,omitempty"`
	Defaults    workflowProjectDefaults `json:"defaults" yaml:"defaults"`
	Steps       any                     `json:"steps" yaml:"steps"`
}

type workflowProjectDefaults struct {
	ContextID     string `json:"contextId,omitempty" yaml:"contextId,omitempty"`
	ContextName   string `json:"contextName,omitempty" yaml:"contextName,omitempty"`
	IncludeOutput *bool  `json:"includeOutput,omitempty" yaml:"includeOutput,omitempty"`
	KeepSession   *bool  `json:"keepSession,omitempty" yaml:"keepSession,omitempty"`
}

type workflowNode interface {
	isWorkflowNode()
}

type workflowStepNode struct {
	Step workflowStep
}

type workflowParallelNode struct {
	Steps []workflowStep
}

func (workflowStepNode) isWorkflowNode()     {}
func (workflowParallelNode) isWorkflowNode() {}

type workflowStep struct {
	Name      string
	File      string
	Code      string
	Log       string
	Listing   string
	Variables map[string]string
}

type workflowDocument struct {
	Path        string
	Dir         string
	Config      workflowProjectConfig
	Nodes       []workflowNode
	User        workflowUserConfig
	MaxParallel int
}

type workflowRunResult struct {
	OK          bool                 `json:"ok"`
	Name        string               `json:"name,omitempty"`
	Path        string               `json:"path,omitempty"`
	ContextID   string               `json:"contextId,omitempty"`
	ContextName string               `json:"contextName,omitempty"`
	SessionID   string               `json:"sessionId,omitempty"`
	Steps       []workflowNodeResult `json:"steps,omitempty"`
	Warnings    []string             `json:"warnings,omitempty"`
	Error       string               `json:"error,omitempty"`
}

type workflowNodeResult struct {
	Kind             string               `json:"kind,omitempty"`
	Name             string               `json:"name,omitempty"`
	File             string               `json:"file,omitempty"`
	Path             string               `json:"path,omitempty"`
	SessionID        string               `json:"sessionId,omitempty"`
	JobID            string               `json:"jobId,omitempty"`
	State            string               `json:"state,omitempty"`
	JobConditionCode int                  `json:"jobConditionCode,omitempty"`
	LogPath          string               `json:"logPath,omitempty"`
	ListingPath      string               `json:"listingPath,omitempty"`
	Warnings         []string             `json:"warnings,omitempty"`
	Error            string               `json:"error,omitempty"`
	Children         []workflowNodeResult `json:"children,omitempty"`
}

func newWorkflowCommand(ioStreams cliIO) *cobra.Command {
	opts := &workflowOptions{
		cfg: cliConfig{
			Timeout:      5 * time.Minute,
			PollInterval: 2 * time.Second,
		},
		maxParallel: defaultWorkflowMaxParallel,
	}

	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Run, validate, and inspect workflow plans",
	}
	addConfigFlags(cmd.PersistentFlags(), &opts.cfg)
	cmd.PersistentFlags().StringVar(&opts.cfg.ContextID, "context-id", "", "Compute context ID")
	cmd.PersistentFlags().StringVar(&opts.cfg.ContextName, "context-name", "", "Compute context name")
	cmd.PersistentFlags().DurationVar(&opts.cfg.PollInterval, "poll-interval", opts.cfg.PollInterval, "Compute job polling interval")
	cmd.PersistentFlags().BoolVar(&opts.cfg.KeepSession, "keep-session", opts.cfg.KeepSession, "Keep the Compute session after the workflow completes")
	cmd.PersistentFlags().BoolVar(&opts.cfg.IncludeOutput, "include-output", opts.cfg.IncludeOutput, "Include log and listing text in JSON output")
	cmd.PersistentFlags().IntVar(&opts.maxParallel, "max-parallel", opts.maxParallel, "Maximum workflow steps to run at once inside a parallel group")
	cmd.PersistentFlags().StringVar(&opts.userConfigFile, "user-config", "", "Path to a workflow user config file")

	cmd.AddCommand(newWorkflowRunCommand(ioStreams, opts))
	cmd.AddCommand(newWorkflowValidateCommand(ioStreams, opts))
	return cmd
}

func newWorkflowRunCommand(ioStreams cliIO, opts *workflowOptions) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a workflow plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" && len(args) > 0 {
				file = args[0]
			}
			if err := requireFlag("file", file); err != nil {
				return writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, workflowRunResult{Error: err.Error()})
			}
			overrides := workflowFlagOverrides{
				includeOutput: cmd.Flag("include-output").Changed,
				keepSession:   cmd.Flag("keep-session").Changed,
				maxParallel:   cmd.Flag("max-parallel").Changed,
				contextID:     cmd.Flag("context-id").Changed,
				contextName:   cmd.Flag("context-name").Changed,
			}
			result, err := runWorkflow(ioStreams, *opts, file, overrides)
			if err != nil {
				return err
			}
			return writeWorkflowOutput(ioStreams.stdout, opts.cfg.Output, result)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Path to a workflow file in JSON or YAML format")
	return cmd
}

func newWorkflowValidateCommand(ioStreams cliIO, opts *workflowOptions) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a workflow file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" && len(args) > 0 {
				file = args[0]
			}
			if err := requireFlag("file", file); err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			doc, err := loadWorkflowDocument(file)
			if err != nil {
				return writeCommandFailure(ioStreams.stdout, opts.cfg.Output, err)
			}
			message := fmt.Sprintf("workflow %q is valid", doc.Config.Name)
			if doc.Config.Name == "" {
				message = fmt.Sprintf("workflow file %q is valid", file)
			}
			return writeCommandOutput(ioStreams.stdout, opts.cfg.Output, commandResult{OK: true, Data: message})
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Path to a workflow file in JSON or YAML format")
	return cmd
}

func runWorkflow(ioStreams cliIO, opts workflowOptions, file string, overrides workflowFlagOverrides) (workflowRunResult, error) {
	doc, err := loadWorkflowDocument(file)
	if err != nil {
		return workflowRunResult{Error: err.Error()}, writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, workflowRunResult{Error: err.Error()})
	}

	userConfig, err := loadWorkflowUserConfig(opts.cfg, opts.userConfigFile)
	if err != nil {
		return workflowRunResult{Error: err.Error()}, writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, workflowRunResult{Error: err.Error()})
	}
	doc.User = userConfig

	cfg := opts.cfg
	if !overrides.includeOutput && doc.Config.Defaults.IncludeOutput != nil {
		cfg.IncludeOutput = *doc.Config.Defaults.IncludeOutput
	}
	if !overrides.keepSession && doc.Config.Defaults.KeepSession != nil {
		cfg.KeepSession = *doc.Config.Defaults.KeepSession
	}

	maxParallel := opts.maxParallel
	if maxParallel <= 0 {
		maxParallel = defaultWorkflowMaxParallel
	}
	if doc.Config.MaxParallel > 0 && !overrides.maxParallel {
		maxParallel = doc.Config.MaxParallel
	}

	client, ctx, cancel, cfg, err := newConfiguredClient(cfg)
	if err != nil {
		return workflowRunResult{Error: err.Error()}, writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, workflowRunResult{Error: err.Error()})
	}
	defer cancel()

	contextID, contextName, err := resolveWorkflowContext(ctx, client, cfg, doc, overrides)
	if err != nil {
		return workflowRunResult{Error: err.Error()}, writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, workflowRunResult{Error: err.Error()})
	}

	sessionName := doc.Config.Name
	if sessionName == "" {
		sessionName = fmt.Sprintf("viya-cli-workflow-%d", time.Now().Unix())
	}
	sessionReq := viya.CreateComputeSessionRequest{
		Version: 3,
		Name:    sessionName,
	}
	if autoexec := strings.TrimSpace(userConfig.Autoexec); autoexec != "" {
		sessionReq.Environment = &viya.ComputeEnvironment{InitCode: splitCodeLines(autoexec)}
	}
	session, err := client.CreateComputeSession(ctx, contextID, sessionReq)
	if err != nil {
		return workflowRunResult{ContextID: contextID, ContextName: contextName, Error: err.Error()}, writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, workflowRunResult{ContextID: contextID, ContextName: contextName, Error: err.Error()})
	}

	result := workflowRunResult{
		OK:          true,
		Name:        doc.Config.Name,
		Path:        doc.Path,
		ContextID:   contextID,
		ContextName: contextName,
		SessionID:   session.ID,
	}

	if !cfg.KeepSession {
		defer func() {
			_ = client.DeleteComputeSession(context.Background(), session.ID)
		}()
	}

	runner := workflowRunner{
		client:      client,
		ctx:         ctx,
		cfg:         cfg,
		doc:         doc,
		userConfig:  userConfig,
		sessionID:   session.ID,
		maxParallel: maxParallel,
	}
	steps, err := runner.run(ctx, doc.Nodes)
	result.Steps = steps
	if err != nil {
		result.OK = false
		result.Error = err.Error()
		return result, writeWorkflowFailure(ioStreams.stdout, opts.cfg.Output, result)
	}

	return result, nil
}

type computeJobOptions struct {
	IncludeOutput bool
	PollInterval  time.Duration
}

type computeJobResult struct {
	JobID            string   `json:"jobId,omitempty"`
	State            string   `json:"state,omitempty"`
	JobConditionCode int      `json:"jobConditionCode,omitempty"`
	Log              string   `json:"log,omitempty"`
	Listing          string   `json:"listing,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
}

func runComputeJob(ctx context.Context, client *viya.Client, sessionID string, req viya.CreateComputeJobRequest, opts computeJobOptions) (computeJobResult, error) {
	job, err := client.CreateComputeJob(ctx, sessionID, req)
	if err != nil {
		return computeJobResult{}, err
	}
	result := computeJobResult{JobID: job.ID}
	state, err := waitComputeJob(ctx, client, sessionID, job.ID, opts.PollInterval)
	if err != nil {
		result.State = state
		return result, err
	}
	result.State = state
	if jobInfo, err := client.GetComputeJobInfo(ctx, sessionID, job.ID); err == nil {
		result.JobConditionCode = jobInfo.JobConditionCode
	}
	if opts.IncludeOutput {
		if result.Log, err = client.GetComputeJobLogText(ctx, sessionID, job.ID); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("get log: %v", err))
		}
		if result.Listing, err = client.GetComputeJobListingText(ctx, sessionID, job.ID); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("get listing: %v", err))
		}
	}
	return result, nil
}

type workflowRunner struct {
	client      *viya.Client
	ctx         context.Context
	cfg         cliConfig
	doc         workflowDocument
	userConfig  workflowUserConfig
	sessionID   string
	maxParallel int
}

func (r workflowRunner) run(ctx context.Context, nodes []workflowNode) ([]workflowNodeResult, error) {
	results := make([]workflowNodeResult, 0, len(nodes))
	for _, node := range nodes {
		switch n := node.(type) {
		case workflowStepNode:
			stepResult, err := r.runStep(ctx, n.Step)
			results = append(results, stepResult)
			if err != nil {
				return results, err
			}
		case workflowParallelNode:
			parallelResult, err := r.runParallelGroup(ctx, n.Steps)
			results = append(results, parallelResult)
			if err != nil {
				return results, err
			}
		default:
			return results, fmt.Errorf("unsupported workflow node %T", node)
		}
	}
	return results, nil
}

func (r workflowRunner) runParallelGroup(ctx context.Context, steps []workflowStep) (workflowNodeResult, error) {
	groupCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]workflowNodeResult, len(steps))
	maxParallel := r.maxParallel
	if maxParallel <= 0 {
		maxParallel = defaultWorkflowMaxParallel
	}
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	for i, step := range steps {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-groupCtx.Done():
				results[i] = workflowNodeResult{Kind: "step", Name: stepDisplayName(step, i), Error: groupCtx.Err().Error()}
				return
			}
			defer func() { <-sem }()

			stepResult, err := r.runStep(groupCtx, step)
			results[i] = stepResult
			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
			}
		}()
	}

	wg.Wait()
	return workflowNodeResult{Kind: "parallel", Children: results}, firstErr
}

func (r workflowRunner) runStep(ctx context.Context, step workflowStep) (workflowNodeResult, error) {
	resolvedFile := ""
	content := ""
	if step.File != "" {
		var err error
		resolvedFile, err = resolveWorkflowPath(r.doc.Dir, step.File)
		if err != nil {
			return workflowNodeResult{Kind: "step", Name: stepDisplayName(step, 0), File: step.File, Error: err.Error()}, err
		}
		fileContent, err := os.ReadFile(resolvedFile)
		if err != nil {
			return workflowNodeResult{Kind: "step", Name: stepDisplayName(step, 0), File: step.File, Path: resolvedFile, Error: err.Error()}, err
		}
		content = string(fileContent)
	}

	jobCode := buildWorkflowJobCode(r.userConfig, step, content, r.doc, resolvedFile)
	jobReq := viya.CreateComputeJobRequest{
		Version:   3,
		Name:      stepDisplayName(step, 0),
		Code:      splitCodeLines(jobCode),
		Variables: workflowVariables(r.doc, step, resolvedFile),
	}

	includeOutput := workflowIncludeOutput(r.cfg, r.doc.Config.Defaults.IncludeOutput)
	if step.Log != "" || step.Listing != "" {
		includeOutput = true
	}
	result, err := runComputeJob(ctx, r.client, r.sessionID, jobReq, computeJobOptions{IncludeOutput: includeOutput, PollInterval: r.cfg.PollInterval})
	stepResult := workflowNodeResult{
		Kind:             "step",
		Name:             stepDisplayName(step, 0),
		File:             step.File,
		Path:             resolvedFile,
		SessionID:        r.sessionID,
		JobID:            result.JobID,
		State:            result.State,
		JobConditionCode: result.JobConditionCode,
		Warnings:         append([]string(nil), result.Warnings...),
	}
	if err != nil {
		stepResult.Error = err.Error()
		return stepResult, err
	}
	if result.State != "completed" {
		err = fmt.Errorf("compute job finished with state %q", result.State)
		stepResult.Error = err.Error()
		return stepResult, err
	}

	if step.Log != "" && result.Log != "" {
		logPath, writeErr := writeWorkflowArtifact(r.doc.Dir, step.Log, result.Log)
		if writeErr != nil {
			stepResult.Warnings = append(stepResult.Warnings, fmt.Sprintf("write log: %v", writeErr))
		} else {
			stepResult.LogPath = logPath
		}
	}
	if step.Listing != "" && result.Listing != "" {
		listingPath, writeErr := writeWorkflowArtifact(r.doc.Dir, step.Listing, result.Listing)
		if writeErr != nil {
			stepResult.Warnings = append(stepResult.Warnings, fmt.Sprintf("write listing: %v", writeErr))
		} else {
			stepResult.ListingPath = listingPath
		}
	}
	return stepResult, nil
}

func workflowIncludeOutput(cfg cliConfig, defaultValue *bool) bool {
	if cfg.IncludeOutput {
		return true
	}
	if defaultValue != nil {
		return *defaultValue
	}
	return false
}

func workflowVariables(doc workflowDocument, step workflowStep, resolvedFile string) map[string]any {
	vars := map[string]any{
		"WORKFLOW_FILE":         doc.Path,
		"WORKFLOW_DIR":          doc.Dir,
		"WORKFLOW_NAME":         doc.Config.Name,
		"WORKFLOW_STEP_NAME":    stepDisplayName(step, 0),
		"WORKFLOW_STEP_FILE":    step.File,
		"WORKFLOW_STEP_PATH":    resolvedFile,
		"WORKFLOW_STEP_LOG":     step.Log,
		"WORKFLOW_STEP_LISTING": step.Listing,
	}
	if resolvedFile != "" {
		vars["_SASPROGRAMFILE"] = resolvedFile
		vars["_SASPROGRAMDIR"] = filepath.Dir(resolvedFile)
	}
	for key, value := range doc.User.Variables {
		vars[key] = value
	}
	for key, value := range step.Variables {
		vars[key] = value
	}
	return vars
}

func buildWorkflowJobCode(user workflowUserConfig, step workflowStep, fileContent string, doc workflowDocument, resolvedFile string) string {
	return wrapWorkflowCode(user, step, fileContent, doc, resolvedFile)
}

func writeWorkflowMacroVars(builder *strings.Builder, vars map[string]any) {
	keys := slices.Sorted(maps.Keys(vars))
	for _, key := range keys {
		value := vars[key]
		builder.WriteString("%let ")
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(sasQuotedString(fmt.Sprint(value)))
		builder.WriteString(";\n")
	}
}

func wrapWorkflowCode(user workflowUserConfig, step workflowStep, fileContent string, doc workflowDocument, resolvedFile string) string {
	var builder strings.Builder
	writeWorkflowMacroVars(&builder, workflowVariables(doc, step, resolvedFile))
	appendWorkflowSnippet(&builder, user.PreCode)
	appendWorkflowSnippet(&builder, step.Code)
	appendWorkflowSnippet(&builder, fileContent)
	appendWorkflowSnippet(&builder, user.PostCode)
	return builder.String()
}

func appendWorkflowSnippet(builder *strings.Builder, code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}
	builder.WriteString(code)
	if !strings.HasSuffix(code, "\n") {
		builder.WriteString("\n")
	}
}

func sasQuotedString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func workflowDisplayName(step workflowStep, index int) string {
	if strings.TrimSpace(step.Name) != "" {
		return step.Name
	}
	if base := filepath.Base(step.File); base != "." && base != string(filepath.Separator) {
		return base
	}
	return fmt.Sprintf("step-%d", index+1)
}

func stepDisplayName(step workflowStep, index int) string {
	return workflowDisplayName(step, index)
}

func resolveWorkflowContext(ctx context.Context, client *viya.Client, cfg cliConfig, doc workflowDocument, overrides workflowFlagOverrides) (string, string, error) {
	contextID := firstNonEmpty(doc.Config.Defaults.ContextID, doc.User.ContextID)
	if overrides.contextID {
		contextID = firstNonEmpty(cfg.ContextID, contextID)
	}
	contextName := firstNonEmpty(doc.Config.Defaults.ContextName, doc.User.ContextName)
	if overrides.contextName {
		contextName = firstNonEmpty(cfg.ContextName, contextName)
	}
	if contextID != "" {
		if contextName != "" {
			return contextID, contextName, nil
		}
		info, err := client.GetComputeContextInfo(ctx, contextID)
		if err == nil {
			return contextID, info.Name, nil
		}
		return contextID, "", nil
	}
	if contextName != "" {
		contexts, err := client.GetComputeContexts(ctx)
		if err != nil {
			return "", "", err
		}
		for _, item := range contexts.Items {
			if item.Name == contextName || item.ID == contextName {
				return item.ID, item.Name, nil
			}
		}
		return "", "", fmt.Errorf("compute context %q was not found", contextName)
	}
	return resolveComputeContext(ctx, client, cfg)
}

func loadWorkflowDocument(path string) (workflowDocument, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return workflowDocument{}, err
	}

	var raw yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return workflowDocument{}, fmt.Errorf("decode workflow file %q: %w", path, err)
	}
	root, err := workflowMappingNode(&raw)
	if err != nil {
		return workflowDocument{}, fmt.Errorf("parse workflow file %q: %w", path, err)
	}

	config, err := parseWorkflowProjectConfig(root)
	if err != nil {
		return workflowDocument{}, fmt.Errorf("parse workflow file %q: %w", path, err)
	}

	stepsNode := workflowMappingValue(root, "steps")
	nodes, err := parseWorkflowNodes(stepsNode)
	if err != nil {
		return workflowDocument{}, fmt.Errorf("parse workflow file %q: %w", path, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	return workflowDocument{
		Path:        absPath,
		Dir:         filepath.Dir(absPath),
		Config:      config,
		Nodes:       nodes,
		MaxParallel: config.MaxParallel,
	}, nil
}

func loadWorkflowUserConfig(_ cliConfig, override string) (workflowUserConfig, error) {
	if override != "" {
		content, err := os.ReadFile(override)
		if err != nil {
			return workflowUserConfig{}, err
		}
		return parseWorkflowUserConfig(override, content)
	}

	home, _ := os.UserHomeDir()
	baseDir := ""
	if home != "" {
		baseDir = filepath.Join(home, ".viya-cli")
	}
	if baseDir == "" {
		return workflowUserConfig{}, nil
	}
	for _, candidate := range []string{
		filepath.Join(baseDir, "workflow.json"),
		filepath.Join(baseDir, "workflow.yaml"),
		filepath.Join(baseDir, "workflow.yml"),
	} {
		content, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		return parseWorkflowUserConfig(candidate, content)
	}
	return workflowUserConfig{}, nil
}

func parseWorkflowUserConfig(path string, content []byte) (workflowUserConfig, error) {
	var raw yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return workflowUserConfig{}, fmt.Errorf("decode workflow user config %q: %w", path, err)
	}
	root, err := workflowMappingNode(&raw)
	if err != nil {
		return workflowUserConfig{}, fmt.Errorf("parse workflow user config %q: %w", path, err)
	}
	if err := workflowRequireOnlyKnownKeys(root, path, workflowAllowedUserConfigKeys()); err != nil {
		return workflowUserConfig{}, err
	}
	vars, err := workflowStringMap(workflowMappingValue(root, "variables"), path, "variables")
	if err != nil {
		return workflowUserConfig{}, err
	}
	contextID, err := workflowStringValue(root, path, "contextId", "context_id", "computeContextId")
	if err != nil {
		return workflowUserConfig{}, err
	}
	contextName, err := workflowStringValue(root, path, "contextName", "context_name", "computeContextName")
	if err != nil {
		return workflowUserConfig{}, err
	}
	autoexec, err := workflowStringValue(root, path, "autoexec")
	if err != nil {
		return workflowUserConfig{}, err
	}
	preCode, err := workflowStringValue(root, path, "preCode", "pre_code")
	if err != nil {
		return workflowUserConfig{}, err
	}
	postCode, err := workflowStringValue(root, path, "postCode", "post_code")
	if err != nil {
		return workflowUserConfig{}, err
	}
	return workflowUserConfig{
		ContextID:   contextID,
		ContextName: contextName,
		Autoexec:    autoexec,
		PreCode:     preCode,
		PostCode:    postCode,
		Variables:   vars,
	}, nil
}

func parseWorkflowProjectConfig(root *yaml.Node) (workflowProjectConfig, error) {
	if err := workflowRequireOnlyKnownKeys(root, "workflow", workflowAllowedProjectConfigKeys()); err != nil {
		return workflowProjectConfig{}, err
	}
	if versionNode := workflowMappingValue(root, "version"); versionNode != nil {
		version, err := workflowIntValue(versionNode, "version")
		if err != nil {
			return workflowProjectConfig{}, err
		}
		if version != 1 {
			return workflowProjectConfig{}, fmt.Errorf("unsupported workflow version %d: only version 1 is supported", version)
		}
	}
	stepsNode := workflowMappingValue(root, "steps")
	if stepsNode == nil {
		return workflowProjectConfig{}, fmt.Errorf("steps are required")
	}
	if stepsNode.Kind != yaml.SequenceNode {
		return workflowProjectConfig{}, fmt.Errorf("steps must be an array")
	}
	if len(stepsNode.Content) == 0 {
		return workflowProjectConfig{}, fmt.Errorf("steps must contain at least one item")
	}

	defaultsNode := workflowMappingValue(root, "defaults")
	if defaultsNode != nil {
		if defaultsNode.Kind != yaml.MappingNode {
			return workflowProjectConfig{}, fmt.Errorf("defaults must be an object")
		}
		if err := workflowRequireOnlyKnownKeys(defaultsNode, "defaults", workflowAllowedProjectDefaultsKeys()); err != nil {
			return workflowProjectConfig{}, err
		}
	}

	name, err := workflowStringValue(root, "workflow", "name")
	if err != nil {
		return workflowProjectConfig{}, err
	}
	contextID, err := workflowStringValue(root, "workflow", "contextId", "context_id", "computeContextId")
	if err != nil {
		return workflowProjectConfig{}, err
	}
	if contextID == "" {
		contextID, err = workflowStringValue(defaultsNode, "defaults", "contextId", "context_id", "computeContextId")
		if err != nil {
			return workflowProjectConfig{}, err
		}
	}
	contextName, err := workflowStringValue(root, "workflow", "contextName", "context_name", "computeContextName")
	if err != nil {
		return workflowProjectConfig{}, err
	}
	if contextName == "" {
		contextName, err = workflowStringValue(defaultsNode, "defaults", "contextName", "context_name", "computeContextName")
		if err != nil {
			return workflowProjectConfig{}, err
		}
	}
	includeOutput, err := firstBoolValue(root, defaultsNode, "includeOutput", "include_output")
	if err != nil {
		return workflowProjectConfig{}, err
	}
	keepSession, err := firstBoolValue(root, defaultsNode, "keepSession", "keep_session")
	if err != nil {
		return workflowProjectConfig{}, err
	}
	config := workflowProjectConfig{
		Version:     1,
		Name:        name,
		MaxParallel: defaultWorkflowMaxParallel,
		Defaults: workflowProjectDefaults{
			ContextID:     contextID,
			ContextName:   contextName,
			IncludeOutput: includeOutput,
			KeepSession:   keepSession,
		},
	}
	if maxParallelNode := workflowMappingValue(root, "maxParallel"); maxParallelNode != nil {
		maxParallel, err := workflowIntValue(maxParallelNode, "maxParallel")
		if err != nil {
			return workflowProjectConfig{}, err
		}
		if maxParallel < 1 {
			return workflowProjectConfig{}, fmt.Errorf("maxParallel must be at least 1")
		}
		config.MaxParallel = maxParallel
	} else if maxParallelNode = workflowMappingValue(root, "max_parallel"); maxParallelNode != nil {
		maxParallel, err := workflowIntValue(maxParallelNode, "max_parallel")
		if err != nil {
			return workflowProjectConfig{}, err
		}
		if maxParallel < 1 {
			return workflowProjectConfig{}, fmt.Errorf("max_parallel must be at least 1")
		}
		config.MaxParallel = maxParallel
	}
	return config, nil
}

func parseWorkflowNodes(node *yaml.Node) ([]workflowNode, error) {
	if node == nil {
		return nil, fmt.Errorf("steps are required")
	}
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("steps must be an array")
	}
	if len(node.Content) == 0 {
		return nil, fmt.Errorf("steps must contain at least one item")
	}
	result := make([]workflowNode, 0, len(node.Content))
	for _, item := range node.Content {
		switch item.Kind {
		case yaml.MappingNode:
			step, err := parseWorkflowStep(item)
			if err != nil {
				return nil, err
			}
			result = append(result, workflowStepNode{Step: step})
		case yaml.SequenceNode:
			parallel, err := parseWorkflowParallel(item)
			if err != nil {
				return nil, err
			}
			result = append(result, workflowParallelNode{Steps: parallel})
		default:
			return nil, fmt.Errorf("workflow steps must be objects or nested arrays, got %s", yamlKindName(item.Kind))
		}
	}
	return result, nil
}

func parseWorkflowParallel(node *yaml.Node) ([]workflowStep, error) {
	if node == nil {
		return nil, fmt.Errorf("parallel workflow group must not be empty")
	}
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("parallel workflow groups must be arrays")
	}
	if len(node.Content) == 0 {
		return nil, fmt.Errorf("parallel workflow group must contain at least one step")
	}
	result := make([]workflowStep, 0, len(node.Content))
	for _, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("parallel workflow groups must contain step objects, got %s", yamlKindName(item.Kind))
		}
		step, err := parseWorkflowStep(item)
		if err != nil {
			return nil, err
		}
		result = append(result, step)
	}
	return result, nil
}

func parseWorkflowStep(node *yaml.Node) (workflowStep, error) {
	if node == nil {
		return workflowStep{}, fmt.Errorf("workflow step must not be empty")
	}
	if node.Kind != yaml.MappingNode {
		return workflowStep{}, fmt.Errorf("workflow step must be an object")
	}
	if err := workflowRequireOnlyKnownKeys(node, "workflow step", workflowAllowedStepKeys()); err != nil {
		return workflowStep{}, err
	}
	variables, err := workflowStringMap(workflowMappingValue(node, "variables"), "workflow step", "variables")
	if err != nil {
		return workflowStep{}, err
	}
	name, err := workflowStringValue(node, "workflow step", "name", "id")
	if err != nil {
		return workflowStep{}, err
	}
	file, err := workflowStringValue(node, "workflow step", "file", "work", "path")
	if err != nil {
		return workflowStep{}, err
	}
	code, err := workflowStringValue(node, "workflow step", "code")
	if err != nil {
		return workflowStep{}, err
	}
	logPath, err := workflowStringValue(node, "workflow step", "log")
	if err != nil {
		return workflowStep{}, err
	}
	listingPath, err := workflowStringValue(node, "workflow step", "listing")
	if err != nil {
		return workflowStep{}, err
	}
	step := workflowStep{
		Name:      name,
		File:      file,
		Code:      code,
		Log:       logPath,
		Listing:   listingPath,
		Variables: variables,
	}
	if step.File == "" && strings.TrimSpace(step.Code) == "" {
		return workflowStep{}, fmt.Errorf("workflow step must define either file or code")
	}
	if step.Name == "" && step.File != "" {
		step.Name = filepath.Base(step.File)
	}
	return step, nil
}

func workflowAllowedProjectConfigKeys() map[string]struct{} {
	return map[string]struct{}{
		"$schema":            {},
		"version":            {},
		"name":               {},
		"maxParallel":        {},
		"max_parallel":       {},
		"contextId":          {},
		"context_id":         {},
		"computeContextId":   {},
		"contextName":        {},
		"context_name":       {},
		"computeContextName": {},
		"includeOutput":      {},
		"include_output":     {},
		"keepSession":        {},
		"keep_session":       {},
		"defaults":           {},
		"steps":              {},
	}
}

func workflowAllowedProjectDefaultsKeys() map[string]struct{} {
	return map[string]struct{}{
		"contextId":          {},
		"context_id":         {},
		"computeContextId":   {},
		"contextName":        {},
		"context_name":       {},
		"computeContextName": {},
		"includeOutput":      {},
		"include_output":     {},
		"keepSession":        {},
		"keep_session":       {},
	}
}

func workflowAllowedStepKeys() map[string]struct{} {
	return map[string]struct{}{
		"id":        {},
		"name":      {},
		"file":      {},
		"work":      {},
		"path":      {},
		"code":      {},
		"log":       {},
		"listing":   {},
		"variables": {},
	}
}

func workflowAllowedUserConfigKeys() map[string]struct{} {
	return map[string]struct{}{
		"$schema":            {},
		"contextId":          {},
		"context_id":         {},
		"computeContextId":   {},
		"contextName":        {},
		"context_name":       {},
		"computeContextName": {},
		"autoexec":           {},
		"preCode":            {},
		"pre_code":           {},
		"postCode":           {},
		"post_code":          {},
		"variables":          {},
	}
}

func workflowRequireOnlyKnownKeys(node *yaml.Node, label string, allowed map[string]struct{}) error {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("%s must be an object", label)
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		if _, ok := allowed[key.Value]; !ok {
			return fmt.Errorf("%s contains unknown field %q", label, key.Value)
		}
	}
	return nil
}

func workflowMappingNode(node *yaml.Node) (*yaml.Node, error) {
	if node == nil {
		return nil, fmt.Errorf("workflow file is empty")
	}
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return nil, fmt.Errorf("workflow file is empty")
		}
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("workflow file must be a mapping object")
	}
	return node, nil
}

func workflowMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func workflowStringValue(node *yaml.Node, label string, keys ...string) (string, error) {
	for _, key := range keys {
		valueNode := workflowMappingValue(node, key)
		if valueNode == nil {
			continue
		}
		if valueNode.Kind != yaml.ScalarNode || valueNode.ShortTag() != "!!str" {
			return "", fmt.Errorf("%s.%s must be a string", label, key)
		}
		var text string
		if err := valueNode.Decode(&text); err != nil {
			return "", fmt.Errorf("%s.%s must be a string: %w", label, key, err)
		}
		if text != "" {
			return text, nil
		}
	}
	return "", nil
}

func workflowIntValue(node *yaml.Node, field string) (int, error) {
	var value int
	if err := node.Decode(&value); err != nil {
		return 0, fmt.Errorf("%s must be an integer", field)
	}
	return value, nil
}

func firstBoolValue(root *yaml.Node, fallback *yaml.Node, keys ...string) (*bool, error) {
	if value, err := workflowBoolValue(root, "workflow", keys...); value != nil || err != nil {
		return value, err
	}
	return workflowBoolValue(fallback, "defaults", keys...)
}

func workflowBoolValue(node *yaml.Node, label string, keys ...string) (*bool, error) {
	for _, key := range keys {
		valueNode := workflowMappingValue(node, key)
		if valueNode == nil {
			continue
		}
		if valueNode.Kind != yaml.ScalarNode || valueNode.ShortTag() != "!!bool" {
			return nil, fmt.Errorf("%s.%s must be a boolean", label, key)
		}
		var value bool
		if err := valueNode.Decode(&value); err != nil {
			return nil, fmt.Errorf("%s.%s must be a boolean: %w", label, key, err)
		}
		return &value, nil
	}
	return nil, nil
}

func workflowStringMap(node *yaml.Node, label string, field string) (map[string]string, error) {
	if node == nil {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s.%s must be an object", label, field)
	}
	result := make(map[string]string, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		valueNode := node.Content[i+1]
		var value string
		if err := valueNode.Decode(&value); err != nil {
			return nil, fmt.Errorf("%s.%s[%s] must be a string", label, field, key)
		}
		if value != "" {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func yamlKindName(kind yaml.Kind) string {
	switch kind {
	case yaml.DocumentNode:
		return "document"
	case yaml.MappingNode:
		return "mapping"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return "unknown"
	}
}

func resolveWorkflowPath(baseDir, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Clean(filepath.Join(baseDir, path)), nil
}

func writeWorkflowArtifact(baseDir, path string, content string) (string, error) {
	resolved, err := resolveWorkflowPath(baseDir, path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return "", err
	}
	return resolved, nil
}

func writeWorkflowFailure(w io.Writer, output string, result workflowRunResult) error {
	result.OK = false
	if err := writeWorkflowOutput(w, output, result); err != nil {
		return err
	}
	return exitError{code: 1}
}

func writeWorkflowOutput(w io.Writer, output string, result workflowRunResult) error {
	output, err := normalizeOutput(output)
	if err != nil {
		return err
	}
	if output == "json" {
		return writeJSON(w, result)
	}
	return writeWorkflowText(w, result)
}

func writeWorkflowText(w io.Writer, result workflowRunResult) error {
	if !result.OK {
		if result.Error != "" {
			if _, err := fmt.Fprintf(w, "error: %s\n", result.Error); err != nil {
				return err
			}
		}
		return nil
	}
	if result.Name != "" {
		if _, err := fmt.Fprintf(w, "workflow: %s\n", result.Name); err != nil {
			return err
		}
	}
	if result.SessionID != "" {
		if _, err := fmt.Fprintf(w, "session: %s\n", result.SessionID); err != nil {
			return err
		}
	}
	if result.ContextName != "" {
		if _, err := fmt.Fprintf(w, "context: %s\n", result.ContextName); err != nil {
			return err
		}
	}
	for _, step := range result.Steps {
		if err := writeWorkflowNodeText(w, step, 0); err != nil {
			return err
		}
	}
	return nil
}

func writeWorkflowNodeText(w io.Writer, result workflowNodeResult, indent int) error {
	prefix := strings.Repeat("  ", indent)
	if result.Kind == "parallel" {
		if _, err := fmt.Fprintf(w, "%sparallel\n", prefix); err != nil {
			return err
		}
		for _, child := range result.Children {
			if err := writeWorkflowNodeText(w, child, indent+1); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := fmt.Fprintf(w, "%s- %s", prefix, result.Name); err != nil {
		return err
	}
	if result.State != "" {
		if _, err := fmt.Fprintf(w, " [%s]", result.State); err != nil {
			return err
		}
	}
	if result.Error != "" {
		if _, err := fmt.Fprintf(w, " error=%s", result.Error); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}
