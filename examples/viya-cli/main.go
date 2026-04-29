package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dingdayu/go-viya"
)

type staticTokenProvider string

func (p staticTokenProvider) Token(context.Context) (string, error) {
	if p == "" {
		return "", viya.ErrViyaAuthFailed
	}
	return string(p), nil
}

type cliConfig struct {
	BaseURL       string
	ClientID      string
	ClientSecret  string
	Username      string
	Password      string
	AccessToken   string
	ContextID     string
	ContextName   string
	Profile       string
	SASConfigDir  string
	Timeout       time.Duration
	PollInterval  time.Duration
	KeepSession   bool
	IncludeOutput bool
	Output        string
}

type runResult struct {
	OK               bool     `json:"ok"`
	ContextID        string   `json:"contextId,omitempty"`
	ContextName      string   `json:"contextName,omitempty"`
	SessionID        string   `json:"sessionId,omitempty"`
	JobID            string   `json:"jobId,omitempty"`
	State            string   `json:"state,omitempty"`
	JobConditionCode int      `json:"jobConditionCode,omitempty"`
	Log              string   `json:"log,omitempty"`
	Listing          string   `json:"listing,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
	Error            string   `json:"error,omitempty"`
}

type casResult struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type commandResult struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type cliIO struct {
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
}

type runOptions struct {
	code string
	file string
	cfg  cliConfig
}

type casOptions struct {
	cfg cliConfig
}

type dataOptions struct {
	cfg cliConfig
}

type filesOptions struct {
	cfg cliConfig
}

type jobsOptions struct {
	cfg cliConfig
}

type reportsOptions struct {
	cfg cliConfig
}

type dashboardOptions struct {
	cfg cliConfig
}

func main() {
	cmd := newRootCommand(cliIO{
		stdout: os.Stdout,
		stderr: os.Stderr,
		stdin:  os.Stdin,
	})
	cmd.SetArgs(os.Args[1:])

	if err := cmd.Execute(); err != nil {
		var exitErr exitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}
