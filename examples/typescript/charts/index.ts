import {
  BarChart,
  BoxAndWhiskerChart,
  Chart,
  ChartType,
  CompositeChart,
  Daytona,
  LineChart,
  PieChart,
  ScatterChart,
  Image,
} from '@daytonaio/sdk'
import * as fs from 'fs'
import * as path from 'path'

async function main() {
  const daytona = new Daytona()

  //  first, create a sandbox
  const sandbox = await daytona.create(
    {
      image: Image.debianSlim('3.13').pipInstall('matplotlib'),
    },
    {
      onSnapshotCreateLogs: console.log,
    },
  )

  try {
    const response = await sandbox.process.codeRun(code)
    if (response.exitCode !== 0) {
      console.error('Execution failed with exit code', response.exitCode)
      console.error('Output:', response.artifacts?.stdout)
      return
    }
    for (const chart of response.artifacts?.charts || []) {
      saveChartImage(chart)
      printChart(chart)
    }
  } catch (error) {
    console.error('Execution error:', error)
  } finally {
    //  cleanup
    await daytona.delete(sandbox)
  }
}

main()

const code = `
import matplotlib.pyplot as plt
import numpy as np

# Sample data
x = np.linspace(0, 10, 30)
y = np.sin(x)
categories = ['A', 'B', 'C', 'D', 'E']
values = [40, 63, 15, 25, 8]
box_data = [np.random.normal(0, std, 100) for std in range(1, 6)]

# 1. Line Chart
plt.figure(figsize=(8, 5))
plt.plot(x, y, 'b-', linewidth=2)
plt.title('Line Chart')
plt.xlabel('X-axis (seconds)')  # Added unit
plt.ylabel('Y-axis (amplitude)')  # Added unit
plt.grid(True)
plt.show()

# 2. Scatter Plot
plt.figure(figsize=(8, 5))
plt.scatter(x, y, c=y, cmap='viridis', s=100*np.abs(y))
plt.colorbar(label='Value (normalized)')  # Added unit
plt.title('Scatter Plot')
plt.xlabel('X-axis (time in seconds)')  # Added unit
plt.ylabel('Y-axis (signal strength)')  # Added unit
plt.show()

# 3. Bar Chart
plt.figure(figsize=(10, 6))
plt.bar(categories, values, color='skyblue', edgecolor='navy')
plt.title('Bar Chart')
plt.xlabel('Categories')  # No change (categories don't have units)
plt.ylabel('Values (count)')  # Added unit
plt.show()

# 4. Pie Chart
plt.figure(figsize=(8, 8))
plt.pie(values, labels=categories,
        autopct='%1.1f%%',
        colors=plt.cm.Set3.colors, shadow=True, startangle=90)
plt.title('Pie Chart (Distribution in %)')  # Modified title
plt.axis('equal')  # Equal aspect ratio ensures the pie chart is circular
plt.legend()
plt.show()

# 5. Box and Whisker Plot
plt.figure(figsize=(10, 6))
plt.boxplot(box_data, patch_artist=True, 
            boxprops=dict(facecolor='lightblue'),
            medianprops=dict(color='red', linewidth=2))
plt.title('Box and Whisker Plot')
plt.xlabel('Groups (Experiment IDs)')  # Added unit
plt.ylabel('Values (measurement units)')  # Added unit
plt.grid(True, linestyle='--', alpha=0.7)
plt.show()
`

function printChart(chart: Chart) {
  console.log('Type:', chart.type)
  console.log('Title:', chart.title)

  if (chart.type === ChartType.LINE) {
    const lineChart = chart as LineChart
    console.log('X Label:', lineChart.x_label)
    console.log('Y Label:', lineChart.y_label)
    console.log('X Ticks:', lineChart.x_ticks)
    console.log('Y Ticks:', lineChart.y_ticks)
    console.log('X Tick Labels:', lineChart.x_tick_labels)
    console.log('Y Tick Labels:', lineChart.y_tick_labels)
    console.log('X Scale:', lineChart.x_scale)
    console.log('Y Scale:', lineChart.y_scale)
    console.log('Elements:')
    console.dir(lineChart.elements, { depth: null })
  } else if (chart.type === ChartType.SCATTER) {
    const scatterChart = chart as ScatterChart
    console.log('X Label:', scatterChart.x_label)
    console.log('Y Label:', scatterChart.y_label)
    console.log('X Ticks:', scatterChart.x_ticks)
    console.log('Y Ticks:', scatterChart.y_ticks)
    console.log('X Tick Labels:', scatterChart.x_tick_labels)
    console.log('Y Tick Labels:', scatterChart.y_tick_labels)
    console.log('X Scale:', scatterChart.x_scale)
    console.log('Y Scale:', scatterChart.y_scale)
    console.log('Elements:')
    console.dir(scatterChart.elements, { depth: null })
  } else if (chart.type === ChartType.BAR) {
    const barChart = chart as BarChart
    console.log('X Label:', barChart.x_label)
    console.log('Y Label:', barChart.y_label)
    console.log('Elements:', barChart.elements)
  } else if (chart.type === ChartType.PIE) {
    const pieChart = chart as PieChart
    console.log('Elements:', pieChart.elements)
  } else if (chart.type === ChartType.BOX_AND_WHISKER) {
    const boxAndWhiskerChart = chart as BoxAndWhiskerChart
    console.log('X Label:', boxAndWhiskerChart.x_label)
    console.log('Y Label:', boxAndWhiskerChart.y_label)
    console.log('Elements:', boxAndWhiskerChart.elements)
  } else if (chart.type === ChartType.COMPOSITE_CHART) {
    const compositeChart = chart as CompositeChart
    console.log('Elements:\n')
    compositeChart.elements.forEach(printChart)
  }
  console.log()
}

function saveChartImage(chart: Chart) {
  if (!chart.png) {
    console.log('No image data available for this chart')
    return
  }
  const imgData = Buffer.from(chart.png, 'base64')
  const scriptDir = __dirname
  const filename = chart.title
    ? path.join(scriptDir, `${chart.title}.png`)
    : path.join(scriptDir, `chart_${Date.now()}.png`)
  fs.writeFileSync(filename, imgData)
  console.log(`Image saved as: ${filename}`)
}
