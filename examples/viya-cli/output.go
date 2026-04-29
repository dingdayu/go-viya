package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dingdayu/go-viya"
)

func writeFailure(w io.Writer, output string, result runResult) error {
	result.OK = false
	if err := writeRunOutput(w, output, result); err != nil {
		return err
	}
	return exitError{code: 1}
}

func writeCASFailure(w io.Writer, output string, err error) error {
	if writeErr := writeCommandOutput(w, output, commandResult{OK: false, Error: err.Error()}); writeErr != nil {
		return writeErr
	}
	return exitError{code: 1}
}

func writeCommandFailure(w io.Writer, output string, err error) error {
	if writeErr := writeCommandOutput(w, output, commandResult{OK: false, Error: err.Error()}); writeErr != nil {
		return writeErr
	}
	return exitError{code: 1}
}

func writeRunOutput(w io.Writer, output string, result runResult) error {
	output, err := normalizeOutput(output)
	if err != nil {
		return err
	}
	if output == "json" {
		return writeJSON(w, result)
	}
	return writeRunText(w, result)
}

func writeCASOutput(w io.Writer, output string, data any) error {
	return writeCommandOutput(w, output, data)
}

func writeCommandOutput(w io.Writer, output string, data any) error {
	output, err := normalizeOutput(output)
	if err != nil {
		return err
	}
	if output == "json" {
		if result, ok := data.(commandResult); ok {
			return writeJSON(w, result)
		}
		if result, ok := data.(casResult); ok {
			return writeJSON(w, result)
		}
		return writeJSON(w, commandResult{OK: true, Data: data})
	}
	if result, ok := data.(commandResult); ok && !result.OK {
		_, err := fmt.Fprintf(w, "error: %s\n", result.Error)
		return err
	}
	if result, ok := data.(casResult); ok && !result.OK {
		_, err := fmt.Fprintf(w, "error: %s\n", result.Error)
		return err
	}
	return writeCASText(w, data)
}

func normalizeOutput(output string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "", "text":
		return "text", nil
	case "json":
		return "json", nil
	default:
		return "", fmt.Errorf("unsupported output format %q; use text or json", output)
	}
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeRunText(w io.Writer, result runResult) error {
	if !result.OK {
		if result.Error != "" {
			if _, err := fmt.Fprintf(w, "error: %s\n", result.Error); err != nil {
				return err
			}
		}
		if result.State != "" {
			if _, err := fmt.Fprintf(w, "state: %s\n", result.State); err != nil {
				return err
			}
		}
		if result.Log != "" {
			_, err := fmt.Fprintf(w, "\n%s", result.Log)
			return err
		}
		return nil
	}

	if result.Listing != "" {
		_, err := fmt.Fprint(w, result.Listing)
		if err != nil || strings.HasSuffix(result.Listing, "\n") {
			return err
		}
		_, err = fmt.Fprintln(w)
		return err
	}
	if result.Log != "" {
		_, err := fmt.Fprint(w, result.Log)
		if err != nil || strings.HasSuffix(result.Log, "\n") {
			return err
		}
		_, err = fmt.Fprintln(w)
		return err
	}
	_, err := fmt.Fprintf(w, "state: %s\njob: %s\nsession: %s\n", result.State, result.JobID, result.SessionID)
	return err
}

func writeCASText(w io.Writer, data any) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	switch v := data.(type) {
	case viya.CASServersResponse:
		fmt.Fprintln(tw, "NAME\tID\tDESCRIPTION")
		for _, item := range v.Items {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Name, item.ID, item.Description)
		}
	case viya.CASLibsResponse:
		fmt.Fprintln(tw, "NAME\tTYPE\tDESCRIPTION")
		for _, item := range v.Items {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Name, item.Type, item.Description)
		}
	case viya.CASTablesResponse:
		fmt.Fprintln(tw, "NAME\tROWS\tCOLUMNS\tSCOPE")
		for _, item := range v.Items {
			fmt.Fprintf(tw, "%s\t%d\t%d\t%s\n", item.Name, item.RowCount, item.ColumnCount, item.Scope)
		}
	case viya.CASTable:
		fmt.Fprintln(tw, "NAME\tCASLIB\tROWS\tCOLUMNS\tSCOPE")
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n", v.Name, v.CaslibName, v.RowCount, v.ColumnCount, v.Scope)
	case viya.CASTableColumnsResponse:
		fmt.Fprintln(tw, "NAME\tTYPE\tLENGTH\tLABEL\tFORMAT")
		for _, item := range v.Items {
			length := item.Length
			if length == 0 {
				length = item.RawLength
			}
			fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n", item.Name, item.Type, length, item.Label, item.Format)
		}
	case viya.CASTableRowsResponse:
		for i, column := range v.Columns {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, column)
		}
		fmt.Fprintln(tw)
		for _, row := range v.Rows {
			for i, column := range v.Columns {
				if i > 0 {
					fmt.Fprint(tw, "\t")
				}
				fmt.Fprint(tw, row[column])
			}
			fmt.Fprintln(tw)
		}
	case viya.ViyaFilesResponse:
		fmt.Fprintln(tw, "ID\tNAME\tCONTENT TYPE\tSIZE")
		for _, item := range v.Items {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", item.ID, item.Name, item.ContentType, item.Size)
		}
	case viya.ViyaFile:
		fmt.Fprintln(tw, "ID\tNAME\tCONTENT TYPE\tSIZE")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n", v.ID, v.Name, v.ContentType, v.Size)
	case viya.JobExecutionJobsResponse:
		fmt.Fprintln(tw, "ID\tNAME\tSTATE\tCREATED")
		for _, item := range v.Items {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.ID, item.Name, item.State, item.CreationTimeStamp.Format(time.RFC3339))
		}
	case viya.JobExecutionJob:
		fmt.Fprintln(tw, "ID\tNAME\tSTATE\tCREATED")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", v.ID, v.Name, v.State, v.CreationTimeStamp.Format(time.RFC3339))
	case viya.ReportsResponse:
		fmt.Fprintln(tw, "ID\tNAME\tDESCRIPTION\tCREATED BY")
		for _, item := range v.Items {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.ID, item.Name, item.Description, item.CreatedBy)
		}
	case viya.Report:
		fmt.Fprintln(tw, "ID\tNAME\tDESCRIPTION\tCREATED BY")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", v.ID, v.Name, v.Description, v.CreatedBy)
	case viya.ReportImageJob:
		fmt.Fprintln(tw, "ID\tSTATE\tREPORT URI\tCREATED")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", v.ID, v.State, v.ReportURI, v.CreationTimeStamp.Format(time.RFC3339))
	case viya.VisualAnalyticsReportResult:
		fmt.Fprintln(tw, "REPORT ID\tNAME\tURI\tSTATUS")
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", v.ResultReportID, v.ResultReportName, v.ResultReportURI, v.Status)
	case string:
		_, err := fmt.Fprint(w, v)
		if err != nil || strings.HasSuffix(v, "\n") {
			return err
		}
		_, err = fmt.Fprintln(w)
		return err
	case []byte:
		_, err := w.Write(v)
		return err
	default:
		return writeJSON(w, data)
	}
	return tw.Flush()
}
