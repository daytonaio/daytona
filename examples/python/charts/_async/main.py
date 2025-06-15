import asyncio
import base64
import os
import time

from daytona import (
    AsyncDaytona,
    BarChart,
    BoxAndWhiskerChart,
    Chart,
    ChartType,
    CompositeChart,
    CreateSandboxFromImageParams,
    Image,
    LineChart,
    PieChart,
    ScatterChart,
)

code = """
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
plt.xlabel('X-axis (seconds)')
plt.ylabel('Y-axis (amplitude)')
plt.grid(True)
plt.show()

# 2. Scatter Plot
plt.figure(figsize=(8, 5))
plt.scatter(x, y, c=y, cmap='viridis', s=100*np.abs(y))
plt.colorbar(label='Value (normalized)')
plt.title('Scatter Plot')
plt.xlabel('X-axis (time in seconds)')
plt.ylabel('Y-axis (signal strength)')
plt.show()

# 3. Bar Chart
plt.figure(figsize=(10, 6))
plt.bar(categories, values, color='skyblue', edgecolor='navy')
plt.title('Bar Chart')
plt.xlabel('Categories')
plt.ylabel('Values (count)')
plt.show()

# 4. Pie Chart
plt.figure(figsize=(8, 8))
plt.pie(values, labels=categories,
        autopct='%1.1f%%',
        colors=plt.cm.Set3.colors, shadow=True, startangle=90)
plt.title('Pie Chart (Distribution in %)')
plt.axis('equal')
plt.legend()
plt.show()

# 5. Box and Whisker Plot
plt.figure(figsize=(10, 6))
plt.boxplot(box_data, patch_artist=True, 
            boxprops=dict(facecolor='lightblue'),
            medianprops=dict(color='red', linewidth=2))
plt.title('Box and Whisker Plot')
plt.xlabel('Groups (Experiment IDs)')
plt.ylabel('Values (measurement units)')
plt.grid(True, linestyle='--', alpha=0.7)
plt.show()
"""


async def main():
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create(
            CreateSandboxFromImageParams(
                image=Image.debian_slim("3.13").pip_install("matplotlib"),
            ),
            on_snapshot_create_logs=print,
        )
        response = await sandbox.process.code_run(code)

        if response.exit_code != 0:
            print(f"Error: {response.exit_code} {response.result}")
        else:
            for chart in response.artifacts.charts:
                img_data = base64.b64decode(chart.png)
                script_dir = os.path.dirname(os.path.abspath(__file__))
                if chart.title:
                    filename = os.path.join(script_dir, f"{chart.title}.png")
                else:
                    filename = os.path.join(script_dir, f"chart_{time.time()}.png")
                with open(filename, "wb") as f:
                    f.write(img_data)
                print(f"Image saved as: {filename}")

                print_chart(chart)

        await daytona.delete(sandbox)


def print_chart(chart: Chart):
    print(f"Type: {chart.type}")
    print(f"Title: {chart.title}")

    if chart.type == ChartType.LINE and isinstance(chart, LineChart):
        print(f"X Label: {chart.x_label}")
        print(f"Y Label: {chart.y_label}")
        print(f"X Ticks: {chart.x_ticks}")
        print(f"X Tick Labels: {chart.x_tick_labels}")
        print(f"X Scale: {chart.x_scale}")
        print(f"Y Ticks: {chart.y_ticks}")
        print(f"Y Tick Labels: {chart.y_tick_labels}")
        print(f"Y Scale: {chart.y_scale}")
        print("Elements:")
        for element in chart.elements:
            print(f"\n\tLabel: {element.label}")
            print(f"\tPoints: {element.points}")
    elif chart.type == ChartType.SCATTER and isinstance(chart, ScatterChart):
        print(f"X Label: {chart.x_label}")
        print(f"Y Label: {chart.y_label}")
        print(f"X Ticks: {chart.x_ticks}")
        print(f"X Tick Labels: {chart.x_tick_labels}")
        print(f"X Scale: {chart.x_scale}")
        print(f"Y Ticks: {chart.y_ticks}")
        print(f"Y Tick Labels: {chart.y_tick_labels}")
        print(f"Y Scale: {chart.y_scale}")
        print("Elements:")
        for element in chart.elements:
            print(f"\n\tLabel: {element.label}")
            print(f"\tPoints: {element.points}")
    elif chart.type == ChartType.BAR and isinstance(chart, BarChart):
        print(f"X Label: {chart.x_label}")
        print(f"Y Label: {chart.y_label}")
        print("Elements:")
        for element in chart.elements:
            print(f"\n\tLabel: {element.label}")
            print(f"\tGroup: {element.group}")
            print(f"\tValue: {element.value}")
    elif chart.type == ChartType.PIE and isinstance(chart, PieChart):
        print("Elements:")
        for element in chart.elements:
            print(f"\n\tLabel: {element.label}")
            print(f"\tAngle: {element.angle}")
            print(f"\tRadius: {element.radius}")
            print(f"\tAutopct: {element.autopct}")
    elif chart.type == ChartType.BOX_AND_WHISKER and isinstance(chart, BoxAndWhiskerChart):
        print(f"X Label: {chart.x_label}")
        print(f"Y Label: {chart.y_label}")
        print("Elements:")
        for element in chart.elements:
            print(f"\n\tLabel: {element.label}")
            print(f"\tMin: {element.min}")
            print(f"\tFirst Quartile: {element.first_quartile}")
            print(f"\tMedian: {element.median}")
            print(f"\tThird Quartile: {element.third_quartile}")
            print(f"\tMax: {element.max}")
            print(f"\tOutliers: {element.outliers}")
    elif chart.type == ChartType.COMPOSITE_CHART and isinstance(chart, CompositeChart):
        print("Elements:\n")
        for element in chart.elements:
            print_chart(element)
    print()


if __name__ == "__main__":
    asyncio.run(main())
