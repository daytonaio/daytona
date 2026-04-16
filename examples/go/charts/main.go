package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

const code = `
import matplotlib.pyplot as plt
import numpy as np

x = np.linspace(0, 10, 30)
y = np.sin(x)

plt.figure(figsize=(8, 5))
plt.plot(x, y, 'b-', linewidth=2)
plt.title('Line Chart')
plt.xlabel('X-axis')
plt.ylabel('Y-axis')
plt.grid(True)
plt.show()

plt.figure(figsize=(8, 5))
plt.scatter(x, y, c=y, cmap='viridis', s=100*np.abs(y))
plt.colorbar(label='Value')
plt.title('Scatter Plot')
plt.xlabel('X-axis')
plt.ylabel('Y-axis')
plt.show()

categories = ['A', 'B', 'C', 'D', 'E']
values = [40, 63, 15, 25, 8]
plt.figure(figsize=(10, 6))
plt.bar(categories, values, color='skyblue', edgecolor='navy')
plt.title('Bar Chart')
plt.xlabel('Categories')
plt.ylabel('Values')
plt.show()
`

func main() {
	ctx := context.Background()

	client, err := daytona.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	pyVersion := "3.13"
	sandbox, err := client.Create(ctx, types.ImageParams{
		Image: daytona.DebianSlim(&pyVersion).PipInstall([]string{"matplotlib", "numpy"}),
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}, options.WithTimeout(300*time.Second))
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	defer sandbox.Delete(ctx)

	log.Printf("Sandbox created: %s", sandbox.ID)

	resp, err := sandbox.Process.CodeRun(ctx, code)
	if err != nil {
		log.Fatalf("CodeRun failed: %v", err)
	}

	log.Printf("Exit code: %d", resp.ExitCode)

	if resp.Artifacts == nil || len(resp.Artifacts.Charts) == 0 {
		log.Println("No charts found")
		return
	}

	outputDir, _ := os.Getwd()
	for _, chart := range resp.Artifacts.Charts {
		chartType := deref(chart.Type)
		chartTitle := deref(chart.Title)
		log.Printf("Chart type: %s, title: %s, elements: %d", chartType, chartTitle, len(chart.Elements))

		if png := deref(chart.Png); png != "" {
			filename := strings.ReplaceAll(chartTitle, " ", "_") + ".png"
			dest := filepath.Join(outputDir, filename)
			data, err := base64.StdEncoding.DecodeString(png)
			if err != nil {
				log.Printf("Failed to decode PNG: %v", err)
				continue
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				log.Printf("Failed to write: %v", err)
				continue
			}
			log.Printf("Saved chart: %s", dest)
		}

		switch chartType {
		case "line", "scatter":
			fmt.Printf("  X Label: %s\n", deref(chart.XLabel))
			fmt.Printf("  Y Label: %s\n", deref(chart.YLabel))
			if len(chart.Elements) > 0 {
				fmt.Printf("  First element: label=%s, points=%d\n", deref(chart.Elements[0].Label), len(chart.Elements[0].Points))
			}
		case "bar":
			fmt.Printf("  X Label: %s\n", deref(chart.XLabel))
			fmt.Printf("  Y Label: %s\n", deref(chart.YLabel))
			if len(chart.Elements) > 0 {
				fmt.Printf("  First element: label=%s, group=%s\n", deref(chart.Elements[0].Label), deref(chart.Elements[0].Group))
			}
		}
	}

	log.Printf("Total charts: %d", len(resp.Artifacts.Charts))
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
