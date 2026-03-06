// cmd/azure-devops-mcp/main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/markistaylor/azure-devops-mcp/internal/controller"
)

func main() {
	cfg := controller.Config{
		OrgURL:  os.Getenv("AZURE_DEVOPS_ORG_URL"),
		PAT:     os.Getenv("AZURE_DEVOPS_PAT"),
		Project: os.Getenv("AZURE_DEVOPS_PROJECT"),
	}

	if cfg.OrgURL == "" {
		log.Fatal("AZURE_DEVOPS_ORG_URL is required")
	}

	if cfg.PAT == "" {
		log.Fatal("AZURE_DEVOPS_PAT is required")
	}

	if cfg.Project == "" {
		log.Fatal("AZURE_DEVOPS_PROJECT is required")
	}

	if err := controller.Run(context.Background(), cfg); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
