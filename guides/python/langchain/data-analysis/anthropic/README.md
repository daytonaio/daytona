# LangChain Data Analysis Example (LangChain + Daytona)

## Overview

This example demonstrates how to build a [LangChain](https://www.langchain.com/) agent that performs secure data analysis using [Daytona](https://daytona.io) sandboxes. The agent uses the `DaytonaDataAnalysisTool` to execute Python code in an isolated environment, enabling automated data analysis workflows with natural language prompts.

In this example, the agent analyzes a vehicle valuations dataset to understand how vehicle prices vary by manufacturing year and generates a line chart showing average price per year.

## Features

- **Secure sandbox execution:** All Python code runs in isolated Daytona sandboxes
- **Natural language interface:** Describe your analysis task in plain English
- **Automatic artifact handling:** Charts and outputs are automatically captured and saved
- **Multi-step reasoning:** Agent breaks down complex analysis into logical steps
- **File management:** Upload datasets, download results, and manage sandbox files
- **Custom result handlers:** Process execution artifacts (charts, logs) as needed

## Requirements

- **Python:** Version 3.10 or higher is required (for LangChain 1.0+ syntax)

> [!TIP]
> It's recommended to use a virtual environment (`venv` or `poetry`) to isolate project dependencies.

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `ANTHROPIC_API_KEY`: Required for Claude AI model access. Get it from [Anthropic Console](https://console.anthropic.com/)

See the `.env.example` file for the exact structure. Copy `.env.example` to `.env` and fill in your API keys before running.

## Getting Started

Before proceeding, complete the following steps:

1. Ensure Python 3.10 or higher is installed
2. Copy `.env.example` to `.env` and add your API keys
3. Download the dataset (instructions below)

### Setup and Run

1. Create and activate a virtual environment:

   ```bash
   python3.10 -m venv venv  
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:

   ```bash
   pip install -U langchain langchain-anthropic langchain-daytona-data-analysis python-dotenv
   ```

3. Download the dataset:

   ```bash
   curl -o dataset.csv https://download.daytona.io/dataset.csv
   ```

   Or download manually from [https://download.daytona.io/dataset.csv](https://download.daytona.io/dataset.csv) and save as `dataset.csv`

4. Run the example:

   ```bash
   python data_analysis.py
   ```

## Configuration

- **Analysis Prompt:** The main prompt is configured in the `agent.invoke()` call inside `data_analysis.py`. You can modify this prompt to analyze different aspects of the data or try different visualization types.

- **Dataset Description:** When uploading the dataset, provide a clear description of the columns and data cleaning instructions to help the agent understand how to process the data.

- **Result Handler:** The `process_data_analysis_result()` function processes execution artifacts. You can customize this to handle different output types (charts, tables, logs, etc.).

## How It Works

When you run the example, the agent follows this workflow:

1. **Dataset Upload:** The CSV file is uploaded to the Daytona sandbox with metadata describing its structure
2. **Agent Reasoning:** The agent receives your natural language request and plans the analysis steps
3. **Code Generation:** Agent generates Python code to explore, clean, and analyze the data
4. **Sandbox Execution:** Code runs securely in the Daytona sandbox environment
5. **Artifact Processing:** Charts and outputs are captured and processed by your custom handler
6. **Cleanup:** Sandbox resources are automatically cleaned up

You provide the data and describe what insights you need - the agent handles the rest.

## Example Output

When the agent completes the analysis, you'll see output like:

```
Result stdout Original dataset shape: (100000, 15)
After removing missing values: (100000, 15)
After removing non-numeric values: (99946, 15)
After removing year outliers: (96598, 15)
After removing price outliers: (90095, 15)

Cleaned data summary:
               year  price_in_euro
count  90095.000000   90095.000000
mean    2016.698563   22422.266707
std        4.457647   12964.727116
min     2005.000000     150.000000
25%     2014.000000   12980.000000
50%     2018.000000   19900.000000
75%     2020.000000   29500.000000
max     2023.000000   62090.000000

Average price by year:
year
2005.0     5968.124319
2006.0     6870.881523
2007.0     8015.234473
2008.0     8788.644495
2009.0     8406.198576
2010.0    10378.815972
2011.0    11540.640435
2012.0    13306.642261
2013.0    14512.707025
2014.0    15997.682899
2015.0    18563.864358
2016.0    20124.556294
2017.0    22268.083322
2018.0    24241.123673
2019.0    26757.469111
2020.0    29400.163494
2021.0    30720.168646
2022.0    33861.717552
2023.0    33119.840175
Name: price_in_euro, dtype: float64

Total number of vehicles analyzed: 90095
Year range: 2005 - 2023
Price range: €150.00 - €62090.00
Overall average price: €22422.27

Chart saved to chart-0.png
```

The agent generates a professional line chart showing how average vehicle prices increased from 2005 (€5,968) to 2022 (€33,862), with a slight decrease in 2023. The chart is saved as `chart-0.png` in your project directory.

## API Reference

The `DaytonaDataAnalysisTool` provides these key methods:

### download_file

```python
def download_file(remote_path: str) -> bytes
```

Downloads a file from the sandbox by its remote path.

### upload_file

```python
def upload_file(file: IO, description: str) -> SandboxUploadedFile
```

Uploads a file to the sandbox with a description of its structure and contents.

### install_python_packages

```python
def install_python_packages(package_names: str | list[str]) -> None
```

Installs Python packages in the sandbox using pip.

### close

```python
def close() -> None
```

Closes and deletes the sandbox environment. Always call this when finished to clean up resources.

For the complete API reference and additional methods, see the [documentation](https://www.daytona.io/docs/en/langchain-data-analysis/#10-api-reference).

## License

See the main project LICENSE file for details.

## References

- [LangChain](https://docs.langchain.com/)
- [Daytona](https://daytona.io)
