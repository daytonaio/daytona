# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
import re
import urllib.request
from collections import defaultdict
from pathlib import Path

import dspy
import matplotlib.pyplot as plt
import numpy as np
from daytona_interpreter import DaytonaInterpreter
from dotenv import load_dotenv

load_dotenv(override=True)

logging.basicConfig(level=logging.INFO)

# ── Fetch and split the novel ────────────────────────────────────────────────

GUTENBERG_URL = "https://www.gutenberg.org/cache/epub/1184/pg1184.txt"
CACHE_PATH = Path(__file__).parent / "monte_cristo.txt"


def fetch_chapters() -> list[str]:
    """Download The Count of Monte Cristo and split into chapters."""
    if CACHE_PATH.exists():
        text = CACHE_PATH.read_text(encoding="utf-8")
    else:
        with urllib.request.urlopen(GUTENBERG_URL) as resp:
            text = resp.read().decode("utf-8")
        CACHE_PATH.write_text(text, encoding="utf-8")

    # Strip Project Gutenberg header/footer.
    # "Chapter 1." appears twice: once in the table of contents and once
    # at the actual start of the narrative.  Skip the TOC by searching
    # from after the first occurrence.
    first = text.find("Chapter 1.")
    start = text.find("Chapter 1.", first + 1)
    if start == -1:
        start = first  # fallback if only one occurrence
    end = text.rfind("*** END OF THE PROJECT GUTENBERG EBOOK")
    if end == -1:
        end = len(text)
    text = text[start:end]

    # Split on chapter headings like " Chapter 1. Marseilles—The Arrival"
    parts = re.split(r"(?:^|\n)\s*Chapter \d+\.", text)
    # First element is empty (before Chapter 1), rest are chapter bodies
    # Re-attach chapter numbers for context
    chapter_list = []
    for i, body in enumerate(parts[1:], start=1):
        chapter_list.append(f"Chapter {i}.{body}")

    return chapter_list


# ── Configure DSPy ───────────────────────────────────────────────────────────

lm = dspy.LM("openrouter/google/gemini-3-flash-preview")
dspy.configure(lm=lm)

# ── Run RLM analysis ────────────────────────────────────────────────────────

interpreter = DaytonaInterpreter()

rlm = dspy.RLM(
    signature="chapters: list[str], task: str -> wealth_data: list[dict]",
    interpreter=interpreter,
    max_iterations=40,
    max_llm_calls=500,
    verbose=True,
)

chapters = fetch_chapters()
print(f"Fetched {len(chapters)} chapters")

TASK = (
    "Analyze the economic trajectory of each major character across the novel. "
    "For each chapter where a character's wealth status is mentioned or implied, "
    "produce a dict with keys: chapter (int), character (str), wealth (int 1-10 "
    "where 1=destitute and 10=richest in Paris), and event (str, brief description "
    "of what changed). Track the following characters: Dantès, Danglars, Fernand/"
    "Morcerf, Villefort, and Mercédès. You need to cover each chapter in the book."
)

WEALTH_DATA_PATH = Path(__file__).parent / "wealth_data.json"

try:
    result = rlm(chapters=chapters, task=TASK)
    wealth_data = result.wealth_data
    WEALTH_DATA_PATH.write_text(json.dumps(wealth_data, indent=2), encoding="utf-8")
    print(f"Saved {len(wealth_data)} entries to {WEALTH_DATA_PATH}")
except Exception:
    logging.exception("RLM analysis failed")
    raise
finally:
    interpreter.shutdown()

# ── Plot ─────────────────────────────────────────────────────────────────────

# Group by character
series = defaultdict(lambda: ([], []))
for row in wealth_data:
    xs, ys = series[row["character"]]
    xs.append(row["chapter"])
    ys.append(row["wealth"])

print(f"\n{len(wealth_data)} data points across {len(series)} characters")

CHARACTER_COLORS = {
    "Dantès": "#e0e0e0",
    "Danglars": "#f25c78",
    "Fernand/Morcerf": "#f0b840",
    "Villefort": "#58c4a7",
    "Mercédès": "#b48eed",
}


def smooth(values, window=7):
    if len(values) < window:
        return np.array(values, dtype=float)
    kernel = np.ones(window) / window
    padded = np.pad(values, (window // 2, window // 2), mode="edge")
    return np.convolve(padded, kernel, mode="valid")


plt.style.use("dark_background")
plt.rcParams.update(
    {
        "font.family": "serif",
        "font.size": 14,
        "axes.spines.top": False,
        "axes.spines.right": False,
        "axes.facecolor": "#0d1117",
        "figure.facecolor": "#0d1117",
    }
)

fig, ax = plt.subplots(figsize=(14, 9))

for char, (xs, ys) in sorted(series.items()):
    color = CHARACTER_COLORS.get(char, "#888888")
    ys_arr = np.array(ys, dtype=float)
    ys_smooth = smooth(ys_arr)
    ax.plot(xs, ys_smooth, linewidth=2.2, color=color, label=char, zorder=3)
    ax.scatter(xs, ys_arr, s=8, color=color, alpha=0.25, zorder=2)

ax.set_yticks([1, 10])
ax.set_yticklabels(["Destitute", "Richest\nin Paris"], fontsize=13)
ax.set_ylim(0.5, 10.5)

ax.set_xlim(1, 117)
ax.set_xticks(range(1, 117, 10))
ax.set_xlabel("Chapter", fontsize=15)

ax.yaxis.grid(True, alpha=0.1, linestyle="--", color="#ffffff")
ax.xaxis.grid(False)

ax.set_title("The Count of Monte Cristo — Character Wealth Trajectories", fontsize=18, fontweight="bold", pad=14)

ax.legend(
    loc="upper left",
    frameon=True,
    framealpha=0.6,
    fancybox=True,
    edgecolor="#333333",
    fontsize=13,
    ncol=1,
    bbox_to_anchor=(0.01, 0.93),
)

plt.tight_layout()
plt.savefig("wealth_trajectories.png", dpi=180, bbox_inches="tight")
print("\nSaved wealth_trajectories.png")
plt.show()
