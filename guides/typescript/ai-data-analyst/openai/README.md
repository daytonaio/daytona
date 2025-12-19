# OpenAI Data Analysis Example (OpenAI + Daytona)

## Overview

This example demonstrates how to build a data analysis tool using [OpenAI's API](https://platform.openai.com/) and [Daytona](https://daytona.io) sandboxes. The script executes Python code in an isolated environment to analyze cafe sales data, enabling automated data analysis workflows with natural language prompts.

In this example, the agent analyzes a cafe sales dataset to find the three highest revenue products for January and visualizes the results in a bar chart.

## Features

- **Secure sandbox execution:** All Python code runs in isolated Daytona sandboxes
- **Natural language interface:** Describe your analysis task in plain English
- **Automatic chart generation:** Visualizations are automatically saved as PNG files
- **File handling:** Upload datasets and process results within the sandbox

## Prerequisites

- **Node.js:** Version 18 or higher is required
- **npm:** Included with Node.js installation

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `OPENAI_API_KEY`: Required for OpenAI API access. Get it from [OpenAI Platform](https://platform.openai.com/api-keys)

Create a `.env` file in the project directory with these variables.

## Getting Started

### Setup and Run

1. Install dependencies:

   ```bash
   npm install
   ```

2. Run the example:

   ```bash
   npm run start
   ```

## How It Works

1. An LLM call generates Python code based on the data format and prompt
2. A new Daytona sandbox is created, containing the data file
3. The Python code is executed in the sandbox
4. Any generated charts are saved as PNG files
5. A second LLM call summarizes the code execution results

## Configuration

### Analysis Customization

The main prompt is configured in the `userPrompt` variable in `index.ts`:

```typescript
const userPrompt = `Give the three highest revenue products for the month of January and show them as a bar chart.`;
```

You can modify this to analyze different aspects of the data or try different visualization types.

The example uses `cafe_sales_data.csv`. To use your own dataset, replace this file and update the filename in the script if needed.

### OpenAI Model Configuration

By default, the example uses the following models, as specified in `index.ts`:

```python
const CODING_MODEL = "gpt-5.1"
const SUMMARY_MODEL = "gpt-4o"
```

The coding model is used for high accuracy code generation, and the summary model is used for fast summarization.

See [Models](https://platform.openai.com/docs/models) for all supported models

## Example Output

When the script completes, you'll see output similar to:

```
Prompt: Give the three highest revenue products for the month of January and show them as a bar chart.
Generating code...
Running code...
✓ Chart saved to chart-0.png
Response: Great! It looks like you successfully executed the code and identified the top three revenue-generating products for January:

1. **Matcha Espresso Fusion** with a total revenue of \$2,603.81.
2. **Oat Milk Latte** with a total revenue of \$2,548.65.
3. **Nitro Cold Brew** with a total revenue of \$2,242.41.
```

The chart will be saved as `chart-0.png` in your project directory, showing a bar chart of the top three revenue-generating products for January.

## License

See the main project LICENSE file for details.

## References

- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)
- [Daytona Documentation](https://www.daytona.io/docs)
