package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dingdayu/go-viya"
)

func runCASCommand(ioStreams cliIO, opts *casOptions, call func(context.Context, *viya.Client) (any, error)) error {
	client, ctx, cancel, cfg, err := newConfiguredClient(opts.cfg)
	if err != nil {
		return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
	}
	defer cancel()

	data, err := call(ctx, client)
	if err != nil {
		return writeCASFailure(ioStreams.stdout, opts.cfg.Output, err)
	}

	return writeCASOutput(ioStreams.stdout, cfg.Output, data)
}

func runSAS(ioStreams cliIO, opts *runOptions) error {
	sasCode, err := readCode(opts.code, opts.file, ioStreams.stdin)
	if err != nil {
		return writeFailure(ioStreams.stdout, opts.cfg.Output, runResult{Error: err.Error()})
	}
	if strings.TrimSpace(sasCode) == "" {
		return writeFailure(ioStreams.stdout, opts.cfg.Output, runResult{Error: "SAS code is required; pass --code, --file, or stdin via --file -"})
	}

	client, ctx, cancel, cfg, err := newConfiguredClient(opts.cfg)
	if err != nil {
		return writeFailure(ioStreams.stdout, opts.cfg.Output, runResult{Error: err.Error()})
	}
	defer cancel()

	contextID, contextName, err := resolveComputeContext(ctx, client, cfg)
	if err != nil {
		return writeFailure(ioStreams.stdout, opts.cfg.Output, runResult{Error: err.Error()})
	}

	session, err := client.CreateComputeSession(ctx, contextID, viya.CreateComputeSessionRequest{
		Name: fmt.Sprintf("viya-cli-%d", time.Now().Unix()),
	})
	if err != nil {
		return writeFailure(ioStreams.stdout, opts.cfg.Output, runResult{ContextID: contextID, ContextName: contextName, Error: err.Error()})
	}

	result := runResult{
		ContextID:   contextID,
		ContextName: contextName,
		SessionID:   session.ID,
	}

	if !cfg.KeepSession {
		defer func() {
			_ = client.DeleteComputeSession(context.Background(), session.ID)
		}()
	}

	job, err := client.CreateComputeJob(ctx, session.ID, viya.CreateComputeJobRequest{
		Code:    splitCodeLines(sasCode),
		Version: 3,
	})
	if err != nil {
		result.Error = err.Error()
		return writeFailure(ioStreams.stdout, opts.cfg.Output, result)
	}
	result.JobID = job.ID

	state, err := waitComputeJob(ctx, client, session.ID, job.ID, cfg.PollInterval)
	if err != nil {
		result.State = state
		result.Error = err.Error()
		return writeFailure(ioStreams.stdout, opts.cfg.Output, result)
	}
	result.State = state

	jobInfo, err := client.GetComputeJobInfo(ctx, session.ID, job.ID)
	if err == nil {
		result.JobConditionCode = jobInfo.JobConditionCode
	}

	if cfg.IncludeOutput {
		result.Log, err = client.GetComputeJobLogText(ctx, session.ID, job.ID)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("get log: %v", err))
		}
		result.Listing, err = client.GetComputeJobListingText(ctx, session.ID, job.ID)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("get listing: %v", err))
		}
	}

	result.OK = state == "completed"
	if !result.OK {
		result.Error = fmt.Sprintf("compute job finished with state %q", state)
		return writeFailure(ioStreams.stdout, opts.cfg.Output, result)
	}

	return writeRunOutput(ioStreams.stdout, cfg.Output, result)
}

func readCode(code string, file string, stdin io.Reader) (string, error) {
	if code != "" {
		return code, nil
	}
	if file == "" {
		return "", nil
	}
	if file == "-" {
		content, err := io.ReadAll(stdin)
		return string(content), err
	}
	content, err := os.ReadFile(file)
	return string(content), err
}

func splitCodeLines(code string) []string {
	code = strings.ReplaceAll(code, "\r\n", "\n")
	code = strings.ReplaceAll(code, "\r", "\n")
	lines := strings.Split(code, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func resolveComputeContext(ctx context.Context, client *viya.Client, cfg cliConfig) (id string, name string, err error) {
	if cfg.ContextID != "" {
		if cfg.ContextName != "" {
			return cfg.ContextID, cfg.ContextName, nil
		}
		info, err := client.GetComputeContextInfo(ctx, cfg.ContextID)
		if err == nil {
			return cfg.ContextID, info.Name, nil
		}
		return cfg.ContextID, "", nil
	}

	contexts, err := client.GetComputeContexts(ctx)
	if err != nil {
		return "", "", err
	}
	if len(contexts.Items) == 0 {
		return "", "", fmt.Errorf("no compute contexts were returned")
	}
	if cfg.ContextName != "" {
		for _, item := range contexts.Items {
			if item.Name == cfg.ContextName || item.ID == cfg.ContextName {
				return item.ID, item.Name, nil
			}
		}
		return "", "", fmt.Errorf("compute context %q was not found", cfg.ContextName)
	}
	return contexts.Items[0].ID, contexts.Items[0].Name, nil
}

func waitComputeJob(ctx context.Context, client *viya.Client, sessionID string, jobID string, interval time.Duration) (string, error) {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		state, err := client.GetComputeJobState(ctx, sessionID, jobID)
		if err != nil {
			return state, err
		}
		switch state {
		case "completed", "failed", "canceled", "error":
			return state, nil
		}

		select {
		case <-ctx.Done():
			return state, ctx.Err()
		case <-ticker.C:
		}
	}
}
