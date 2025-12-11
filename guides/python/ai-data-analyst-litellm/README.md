# LiteLLM Data Analysis Example (LiteLLM + Daytona)

## Overview

This example demonstrates how to build a data analysis tool using [LiteLLM](https://litellm.ai/) and [Daytona](https://daytona.io) sandboxes. The script executes Python code in an isolated environment to analyze cafe sales data, enabling automated data analysis workflows with natural language prompts.

In this example, the agent analyzes a cafe sales dataset to find the three highest revenue products for January and visualizes the results in a bar chart.

## Features

- **Multiple LLM Providers:** Easily switch between different LLM providers including Anthropic, OpenAI, Mistral, and more through LiteLLM
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

### Required

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)

### LLM Provider API Keys (choose one based on your provider)

- `ANTHROPIC_API_KEY`: Required if using Anthropic models (default)
- `OPENAI_API_KEY`: Required if using OpenAI models
- `MISTRAL_API_KEY`: Required if using Mistral AI models
- `DEEPSEEK_API_KEY`: Required if using DeepSeek models
- `OPENROUTER_API_KEY`: Required if using OpenRouter models
- See [Providers](https://docs.litellm.ai/docs/providers) for a complete list of providers and required API keys.

Create a `.env` file in the project directory with the appropriate variables for your chosen provider.

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

### Analysis Customization

- **Analysis Prompt:** The main prompt is configured in the `user_prompt` variable in `ai_data_analyst.py`. You can modify this to analyze different aspects of the data or try different visualization types.

- **Dataset:** The example includes `cafe_sales_data.csv`. To use your own dataset, replace this file and update the filename in the script if needed.

### LLM Provider Configuration

By default, the example uses Anthropic's Claude Sonnet 4.0. To switch to a different LLM provider, modify the `model` parameter in `ai_data_analyst.py`:

```python
# Example model configurations (uncomment the one you want to use)
# model = "openai/gpt-4o"
# model = "mistral/mistral-large-latest"
# model = "deepseek/deepseek-chat"
# model = "openrouter/moonshotai/kimi-k2"
model = "anthropic/claude-sonnet-4.0"  # Default
```

See [Providers](https://docs.litellm.ai/docs/providers) for all supported models

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

- [LiteLLM Documentation](https://docs.litellm.ai/docs/)
- [LiteLLM Providers](https://docs.litellm.ai/docs/providers)
- [Daytona Documentation](https://www.daytona.io/docs)
