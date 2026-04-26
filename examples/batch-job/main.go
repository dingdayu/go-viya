package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/dingdayu/go-viya"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	baseURL := mustEnv("VIYA_BASE_URL")
	clientID := mustEnv("VIYA_CLIENT_ID")
	clientSecret := mustEnv("VIYA_CLIENT_SECRET")
	batchContextID := mustEnv("VIYA_BATCH_CONTEXT_ID")

	tokens, err := viya.NewClientCredentialsTokenProvider(baseURL, clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, baseURL, viya.WithTokenProvider(tokens))

	fileSet, err := client.CreateBatchFileSet(ctx, batchContextID)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.DeleteBatchFileSet(context.Background(), fileSet.ID); err != nil {
			log.Printf("delete file set %s: %v", fileSet.ID, err)
		}
	}()

	const programName = "hello.sas"
	program := []byte("data _null_; put 'Hello from go-viya'; run;")
	if err := client.UploadBatchFile(ctx, fileSet.ID, programName, program); err != nil {
		log.Fatal(err)
	}

	job, err := client.CreateBatchJob(ctx, viya.SubmitBatchJobRequest{
		FileSetID:      fileSet.ID,
		SASProgramName: programName,
		WatchOutput:    true,
		Version:        1,
		LauncherOptions: viya.LauncherOptions{
			BatchContextID: batchContextID,
			JobName:        "go-viya-example",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	finalJob, err := client.WaitBatchJobCompleted(ctx, job.ID, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("job %s finished with state=%s returnCode=%d", finalJob.ID, finalJob.State, finalJob.ReturnCode)
}

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
