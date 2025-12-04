# OpenAI Data Analysis Example (OpenAI + Daytona)

## Overview

This example demonstrates how to build a data analysis tool using [OpenAI's API](https://platform.openai.com/) and [Daytona](https://daytona.io) sandboxes. The script executes Python code in an isolated environment to analyze cafe sales data, enabling automated data analysis workflows with natural language prompts.

In this example, the agent analyzes a cafe sales dataset to find the three highest revenue products for January and visualizes the results in a bar chart.

## Features

- **Secure sandbox execution:** All Python code runs in isolated Daytona sandboxes
- **Natural language interface:** Describe your analysis task in plain English
- **Automatic chart generation:** Visualizations are automatically saved as PNG files
- **File handling:** Upload datasets and process results within the sandbox

## Requirements

- **Python:** Version 3.10 or higher is required

> [!TIP]
> It's recommended to use a virtual environment (`venv` or `poetry`) to isolate project dependencies.

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `OPENAI_API_KEY`: Required for OpenAI API access. Get it from [OpenAI Platform](https://platform.openai.com/api-keys)

Create a `.env` file in the project directory with these variables.

## Getting Started

### Setup and Run

1. Create and activate a virtual environment:
   ```bash
   python3.10 -m venv venv  
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:
   ```bash
   pip install -e .
   ```

3. Run the example:
   ```bash
   python ai_data_analyst.py
   ```

## Configuration

- **Analysis Prompt:** The main prompt is configured in the `user_prompt` variable in `ai_data_analyst.py`. You can modify this to analyze different aspects of the data or try different visualization types.

- **Dataset:** The example includes `cafe_sales_data.csv`. To use your own dataset, replace this file and update the filename in the script if needed.

## Example Output

When the script completes, you'll see output similar to:

```
Prompt: Give the three highest revenue products for the month of January and show them as a bar chart.
Generating code...
Running code...
✓ Chart saved to chart-0.png
Response: The analysis is complete. The chart has been saved as chart-0.png and shows the three highest revenue products for January.
```

The chart will be saved as `chart-0.png` in your project directory, showing a bar chart of the top three revenue-generating products for January.

## License

See the main project LICENSE file for details.

## References

- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)
- [Daytona Documentation](https://www.daytona.io/docs)
