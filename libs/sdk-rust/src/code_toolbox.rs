// Copyright 2019-2025 Daytona Labs Inc.
// SPDX-License-Identifier: Apache-2.0

//! Code toolbox implementations for different programming languages.
//!
//! These toolboxes provide language-specific command generation for code execution.

use base64::{engine::general_purpose::STANDARD as BASE64, Engine as _};

/// Parameters for code execution
#[derive(Debug, Clone, Default)]
pub struct CodeRunParams {
    /// Command-line arguments to pass to the code
    pub argv: Vec<String>,
    /// Environment variables to set for the code execution
    pub env: Option<std::collections::HashMap<String, String>>,
}

/// Trait for language-specific code toolboxes
pub trait CodeToolbox: Send + Sync {
    /// Generate a command to run the provided code
    fn get_run_command(&self, code: &str, params: Option<&CodeRunParams>) -> String;
}

/// Python code toolbox implementation
#[derive(Debug, Clone)]
pub struct PythonCodeToolbox;

impl PythonCodeToolbox {
    /// Creates a new Python code toolbox
    pub fn new() -> Self {
        Self
    }

    /// Checks if matplotlib is imported in the given Python code
    fn is_matplotlib_imported(code: &str) -> bool {
        let patterns = [
            r"(?m)^[^#]*import\s+matplotlib",
            r"(?m)^[^#]*from\s+matplotlib",
        ];

        patterns.iter().any(|pattern| {
            regex::Regex::new(pattern)
                .map(|re| re.is_match(code))
                .unwrap_or(false)
        })
    }
}

impl Default for PythonCodeToolbox {
    fn default() -> Self {
        Self::new()
    }
}

impl CodeToolbox for PythonCodeToolbox {
    fn get_run_command(&self, code: &str, params: Option<&CodeRunParams>) -> String {
        let mut base64_code = BASE64.encode(code);

        // If matplotlib is imported, wrap the code with special handling
        if Self::is_matplotlib_imported(code) {
            let code_wrapper = PYTHON_CODE_WRAPPER.replace("{encoded_code}", &base64_code);
            base64_code = BASE64.encode(&code_wrapper);
        }

        let argv = params
            .and_then(|p| {
                if p.argv.is_empty() {
                    None
                } else {
                    Some(&p.argv)
                }
            })
            .map(|argv| argv.join(" "))
            .unwrap_or_default();

        // Execute the bootstrapper code directly with -u flag for unbuffered output
        format!(
            r#"sh -c 'python3 -u -c "exec(__import__(\"base64\").b64decode(\"{}\").decode())" {}'"#,
            base64_code, argv
        )
    }
}

/// TypeScript code toolbox implementation
#[derive(Debug, Clone)]
pub struct TypeScriptCodeToolbox;

impl TypeScriptCodeToolbox {
    /// Creates a new TypeScript code toolbox
    pub fn new() -> Self {
        Self
    }
}

impl Default for TypeScriptCodeToolbox {
    fn default() -> Self {
        Self::new()
    }
}

impl CodeToolbox for TypeScriptCodeToolbox {
    fn get_run_command(&self, code: &str, params: Option<&CodeRunParams>) -> String {
        let base64_code = BASE64.encode(code);
        let argv = params
            .and_then(|p| {
                if p.argv.is_empty() {
                    None
                } else {
                    Some(&p.argv)
                }
            })
            .map(|argv| argv.join(" "))
            .unwrap_or_default();

        format!(
            r#"sh -c 'echo {} | base64 --decode | npx ts-node -O "{{\"module\":\"CommonJS\"}}" -e "$(cat)" x {} 2>&1 | grep -vE "npm notice"'"#,
            base64_code, argv
        )
    }
}

/// JavaScript code toolbox implementation
#[derive(Debug, Clone)]
pub struct JavaScriptCodeToolbox;

impl JavaScriptCodeToolbox {
    /// Creates a new JavaScript code toolbox
    pub fn new() -> Self {
        Self
    }
}

impl Default for JavaScriptCodeToolbox {
    fn default() -> Self {
        Self::new()
    }
}

impl CodeToolbox for JavaScriptCodeToolbox {
    fn get_run_command(&self, code: &str, params: Option<&CodeRunParams>) -> String {
        let base64_code = BASE64.encode(code);
        let argv = params
            .and_then(|p| {
                if p.argv.is_empty() {
                    None
                } else {
                    Some(&p.argv)
                }
            })
            .map(|argv| argv.join(" "))
            .unwrap_or_default();

        format!(
            r#"sh -c 'echo {} | base64 --decode | node -e "$(cat)" {}'"#,
            base64_code, argv
        )
    }
}

/// Supported programming languages
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CodeLanguage {
    Python,
    TypeScript,
    JavaScript,
}

impl CodeLanguage {
    /// Get the default toolbox for this language
    pub fn toolbox(&self) -> Box<dyn CodeToolbox> {
        match self {
            CodeLanguage::Python => Box::new(PythonCodeToolbox::new()),
            CodeLanguage::TypeScript => Box::new(TypeScriptCodeToolbox::new()),
            CodeLanguage::JavaScript => Box::new(JavaScriptCodeToolbox::new()),
        }
    }

    /// Parse a language string into a CodeLanguage
    #[allow(clippy::should_implement_trait)]
    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "python" | "py" => Some(CodeLanguage::Python),
            "typescript" | "ts" => Some(CodeLanguage::TypeScript),
            "javascript" | "js" => Some(CodeLanguage::JavaScript),
            _ => None,
        }
    }
}

// Python code wrapper for matplotlib support
const PYTHON_CODE_WRAPPER: &str = r#"
import base64
import datetime
import hashlib
import io
import json
import linecache
import sys
import traceback
import types
from importlib.abc import Loader, MetaPathFinder
from importlib.util import find_spec, spec_from_loader

# Global variables to hold imported libraries if needed
np = None
mpl = None
pil_img = None


plt_patched = False
processed_figures = set()


def _parse_point(point):
    if isinstance(point, datetime.date):
        return point.isoformat()
    if isinstance(point, np.datetime64):
        return point.astype("datetime64[s]").astype(str)
    return point


def _is_grid_line(line: any) -> bool:
    x_data = line.get_xdata()
    if len(x_data) != 2:
        return False

    y_data = line.get_ydata()
    if len(y_data) != 2:
        return False

    if x_data[0] == x_data[1] or y_data[0] == y_data[1]:
        return True

    return False


def _extract_line_chart_elements(ax):
    elements = []

    for line in ax.get_lines():
        if _is_grid_line(line):
            continue
        label = line.get_label()
        points = [(_parse_point(x), _parse_point(y)) for x, y in zip(line.get_xdata(), line.get_ydata())]

        element = {"label": label, "points": points}
        elements.append(element)

    return elements


def _extract_scatter_chart_elements(ax):
    elements = []

    for collection in ax.collections:
        points = [(_parse_point(x), _parse_point(y)) for x, y in collection.get_offsets()]
        element = {"label": collection.get_label(), "points": points}
        elements.append(element)

    return elements


def _extract_bar_chart_elements(ax):
    elements = []
    change_orientation = False

    for container in ax.containers:
        heights = [rect.get_height() for rect in container]
        if all(height == heights[0] for height in heights):
            # vertical bars
            change_orientation = True
            labels = [label.get_text() for label in ax.get_yticklabels()]
            values = [rect.get_width() for rect in container]
        else:
            # horizontal bars
            labels = [label.get_text() for label in ax.get_xticklabels()]
            values = heights
        for label, value in zip(labels, values):
            element = {"label": label, "group": container.get_label(), "value": value}
            elements.append(element)

    return elements, change_orientation


def _extract_pie_chart_elements(ax):
    elements = []

    wedges = [patch for patch in ax.patches if isinstance(patch, mpl.patches.Wedge)]
    if len(wedges) == 0:
        return elements

    texts = [text_obj.get_text() for text_obj in ax.texts]

    labels = []
    autopcts = []

    if len(texts) == 2 * len(wedges):
        labels = [texts[i] for i in range(0, 2 * len(wedges), 2)]
        autopcts = [texts[i] for i in range(1, 2 * len(wedges), 2)]
    else:
        labels = texts[:len(wedges)]

    for idx, wedge in enumerate(wedges):
        element = {
            "label": labels[idx],
            "angle": abs(wedge.theta2 - wedge.theta1),
            "radius": wedge.r,
            "autopct": autopcts[idx] if autopcts and len(autopcts) > idx else None,
        }
        elements.append(element)

    return elements


# pylint: disable=too-many-branches
def _extract_box_chart_elements(ax):
    change_orientation = False

    xticklabels = [label.get_text() for label in ax.get_xticklabels()]
    boxes = []
    for label, box in zip(xticklabels, ax.patches):
        vertices = box.get_path().vertices
        x_vertices = list(vertices[:, 0])
        y_vertices = list(vertices[:, 1])
        x = min(x_vertices)
        y = min(y_vertices)

        boxes.append(
            {
                "x": x,
                "y": y,
                "label": label,
                "width": max(x_vertices) - x,
                "height": max(y_vertices) - y,
                "outliers": [],
            }
        )

    orientation = "horizontal"
    if all(box["height"] == boxes[0]["height"] for box in boxes):
        orientation = "vertical"

    if orientation == "vertical":
        change_orientation = True
        for box in boxes:
            box["x"], box["y"] = box["y"], box["x"]
            box["width"], box["height"] = box["height"], box["width"]

    for line in ax.lines:
        xdata = line.get_xdata()
        ydata = line.get_ydata()

        if orientation == "vertical":
            xdata, ydata = ydata, xdata

        if len(xdata) <= 1 or len(ydata) != 2:
            continue

        for box in boxes:
            if box["x"] <= xdata[0] <= xdata[1] <= box["x"] + box["width"]:
                # Horizontal line (median or cap)
                if abs(ydata[0] - ydata[1]) < 0.001 and box["y"] <= ydata[0] <= box["y"] + box["height"]:
                    box["median"] = ydata[0]
                # Vertical line (whiskers)
                elif abs(xdata[0] - xdata[1]) < 0.001:
                    y_min = min(ydata)
                    y_max = max(ydata)

                    # If attached to bottom of box
                    if abs(y_max - box["y"]) < 0.001:
                        box["whisker_lower"] = y_min

                    # If attached to top of box
                    elif abs(y_min - (box["y"] + box["height"])) < 0.001:
                        box["whisker_upper"] = y_max
                break

    outlier_candidates = []

    # Check for any markers in all artists
    for artist in ax.get_children():
        if hasattr(artist, "get_xdata") and hasattr(artist, "get_ydata"):
            try:
                xdata = artist.get_xdata()
                ydata = artist.get_ydata()

                if orientation == "vertical":
                    xdata, ydata = ydata, xdata

                if isinstance(xdata, (list, np.ndarray)) and isinstance(ydata, (list, np.ndarray)):
                    for i in range(min(len(xdata), len(ydata))):
                        outlier_candidates.append((float(xdata[i]), float(ydata[i])))
            except:
                pass

    # Assign points to boxes and determine if they're outliers
    for x, y in outlier_candidates:
        for box in boxes:
            if box["x"] <= x <= box["x"] + box["width"]:
                box_center = box["x"] + box["width"] / 2
                if abs(x - box_center) < 0.001:
                    y_min = box["y"]
                    y_max = box["y"] + box["height"]

                    if box.get("whisker_lower", None):
                        y_min = box["whisker_lower"]
                    if box.get("whisker_upper", None):
                        y_max = box["whisker_upper"]
                    if y < y_min or y > y_max:
                        box["outliers"].append(y)
                break

    return [
        {
            "label": box["label"],
            "min": box.get("whisker_lower", None),
            "first_quartile": box["y"],
            "median": box.get("median", None),
            "third_quartile": box["y"] + box["height"],
            "max": box.get("whisker_upper", None),
            "outliers": box["outliers"],
        }
        for box in boxes
    ], change_orientation


def _save_figure_as_base64(fig, bbox_inches="tight", dpi=100):
    # First save with matplotlib
    png_buffer = io.BytesIO()
    fig.savefig(png_buffer, format="png", bbox_inches=bbox_inches, dpi=dpi)
    png_buffer.seek(0)

    # Open with PIL and apply maximum compression
    with pil_img.open(png_buffer) as img:
        optimized_buffer = io.BytesIO()
        img.save(optimized_buffer, format="png", optimize=True, quality=100, compress_level=9)
        optimized_buffer.seek(0)
        return base64.b64encode(optimized_buffer.getvalue()).decode("utf-8")


def _get_figure_hash(fig):
    png_buffer = io.BytesIO()
    fig.savefig(png_buffer, format="png", dpi=50)
    return hashlib.md5(png_buffer.getvalue()).hexdigest()


def _get_chart_type(ax):
    objects = list(
        filter(
            lambda obj: not isinstance(obj, mpl.text.Text) and not isinstance(obj, mpl.patches.Shadow),
            ax._children,  # pylint: disable=protected-access
        )
    )

    # Check for Line plots
    if all(isinstance(line, mpl.lines.Line2D) for line in objects):
        return "line"

    if all(isinstance(box_or_path, (mpl.patches.PathPatch, mpl.lines.Line2D)) for box_or_path in objects):
        return "box_and_whisker"

    filtered = []
    for obj in objects:
        if isinstance(obj, mpl.lines.Line2D) and _is_grid_line(obj):
            continue
        filtered.append(obj)

    objects = filtered

    # Check for Scatter plots
    if all(isinstance(path, mpl.collections.PathCollection) for path in objects):
        return "scatter"

    # Check for Bar plots
    if all(isinstance(rect, mpl.patches.Rectangle) for rect in objects):
        return "bar"

    # Check for Pie plots
    if all(isinstance(artist, mpl.patches.Wedge) for artist in objects):
        return "pie"

    return "unknown"


def _is_auto_empty_axis(ax):
    return ax.get_subplotspec() is not None and not ax.has_data()


def _is_colorbar_axis(ax):
    return any(
        # pylint: disable=protected-access
        isinstance(child, mpl.colorbar._ColorbarSpine)
        for child in ax.get_children()
    )


def _filter_out_unwanted_axes(axes):
    return [ax for ax in axes if not _is_auto_empty_axis(ax) and not _is_colorbar_axis(ax)]


def _extract_ticks_data(converter, ticks) -> list:
    if isinstance(converter, mpl.dates._SwitchableDateConverter):  # pylint: disable=protected-access
        return [mpl.dates.num2date(tick).isoformat() for tick in ticks]
    try:
        return [float(tick) for tick in ticks]
    except Exception:
        return list(ticks)


def _extract_scale(converter, scale: str, ticks, labels) -> str:
    if isinstance(converter, mpl.dates._SwitchableDateConverter):  # pylint: disable=protected-access
        return "datetime"

    # If the scale is not linear, it can't be categorical
    if scale != "linear":
        return scale

    # If all the ticks are integers and are in order from 0 to n-1
    # and the labels aren't corresponding to the ticks, it's categorical
    for i, tick_and_label in enumerate(zip(ticks, labels)):
        tick, label = tick_and_label
        if isinstance(tick, (int, float)) and tick == i and str(i) != label:
            continue
        # Found a tick, which wouldn't be in a categorical scale
        return "linear"

    return "categorical"


def _extract_chart_data(ax):
    data = {}

    data["title"] = ax.get_title()

    data["x_label"] = ax.get_xlabel()
    data["y_label"] = ax.get_ylabel()

    x_tick_labels = [label.get_text() for label in ax.get_xticklabels()]
    data["x_ticks"] = _extract_ticks_data(ax.xaxis.get_converter(), ax.get_xticks())
    data["x_tick_labels"] = x_tick_labels
    data["x_scale"] = _extract_scale(ax.xaxis.get_converter(), ax.get_xscale(), ax.get_xticks(), x_tick_labels)

    y_tick_labels = [label.get_text() for label in ax.get_yticklabels()]
    data["y_ticks"] = _extract_ticks_data(ax.yaxis.get_converter(), ax.get_yticks())
    data["y_tick_labels"] = y_tick_labels
    data["y_scale"] = _extract_scale(ax.yaxis.get_converter(), ax.get_yscale(), ax.get_yticks(), y_tick_labels)

    chart_type = _get_chart_type(ax)
    elements = []
    change_orientation = False

    if chart_type == "line":
        elements = _extract_line_chart_elements(ax)
    elif chart_type == "scatter":
        elements = _extract_scatter_chart_elements(ax)
    elif chart_type == "bar":
        elements, change_orientation = _extract_bar_chart_elements(ax)
    elif chart_type == "box_and_whisker":
        elements, change_orientation = _extract_box_chart_elements(ax)
    elif chart_type == "pie":
        elements = _extract_pie_chart_elements(ax)

    if change_orientation:
        data["x_label"], data["y_label"] = data["y_label"], data["x_label"]

    data["type"] = chart_type
    data["elements"] = elements

    return data


def _custom_json_serializer(obj):
    if isinstance(obj, np.integer):
        return int(obj)
    if isinstance(obj, np.floating):
        return float(obj)
    if isinstance(obj, np.ndarray):
        return obj.tolist()
    if isinstance(obj, set):
        return list(obj)
    raise TypeError(f"Type {type(obj)} not serializable")


def extract_and_print_figure_metadata(fig):
    """Extract metadata from a matplotlib figure and print as JSON"""
    metadata = {}
    subplots = []

    axes = _filter_out_unwanted_axes(fig.axes)

    for ax in axes:
        data = _extract_chart_data(ax)
        subplots.append(data)

    if len(subplots) > 1:
        metadata = {
            "title": fig.texts[0].get_text() if fig.texts and len(fig.texts) > 0 else None,
            "type": "composite_chart",
            "elements": subplots,
        }
    else:
        metadata = subplots[0] if subplots and len(subplots) > 0 else {"type": "unknown"}

    metadata["png"] = _save_figure_as_base64(fig)
    json_output = {"type": "chart", "value": metadata}

    print(f"dtn_artifact_k39fd2:{json.dumps(json_output, default=_custom_json_serializer)}")


class MatplotlibFinder(MetaPathFinder):
    """Custom finder to intercept matplotlib.pyplot imports"""

    def find_spec(self, fullname, path, target=None):  # pylint: disable=unused-argument
        global plt_patched, np, mpl, pil_img  # pylint: disable=global-statement
        if fullname == "matplotlib.pyplot" and not plt_patched:
            plt_patched = True

            # Import numpy and matplotlib once we are sure we need them
            # pylint: disable=import-outside-toplevel
            import matplotlib
            import numpy
            from PIL import Image

            # Store them in global variables for use throughout the module
            np = numpy
            mpl = matplotlib
            pil_img = Image

            original_spec = find_spec(fullname)
            if original_spec is None:
                return None
            return spec_from_loader(
                fullname,
                MatplotlibLoader(original_spec.loader),
                origin=original_spec.origin,
                is_package=original_spec.submodule_search_locations is not None,
            )
        return None


class MatplotlibLoader(Loader):
    """Custom loader to patch the matplotlib.pyplot module"""

    def __init__(self, original_loader):
        self.original_loader = original_loader

    def create_module(self, spec):
        return self.original_loader.create_module(spec)

    def exec_module(self, module):
        self.original_loader.exec_module(module)
        if hasattr(module, "show"):
            original_show = module.show

            def custom_show(*args, **kwargs):
                global processed_figures  # pylint: disable=global-variable-not-assigned
                fig_nums = module.get_fignums()
                for fig_num in fig_nums:
                    fig = module.figure(fig_num)
                    fig_hash = _get_figure_hash(fig)
                    if fig_hash not in processed_figures:
                        extract_and_print_figure_metadata(fig)
                        processed_figures.add(fig_hash)
                result = original_show(*args, **kwargs)
                module.close("all")
                return result

            module.show = custom_show


def setup_user_code_environment(code):
    """Set up the module to run user code in"""
    module = types.ModuleType("__main__")
    module.__file__ = "<target_code>"
    sys.modules["__main__"] = module
    code_lines = code.splitlines()
    linecache.cache["<target_code>"] = (len(code), None, code_lines, "<target_code>")
    return module


def run_user_code(code):
    """Run the user code with the matplotlib interceptor installed"""
    # Install matplotlib interceptor
    sys.meta_path.insert(0, MatplotlibFinder())

    # Set up clean environment for user code
    module = setup_user_code_environment(code)

    # Compile and run the code
    compiled = compile(code, "<target_code>", "exec")

    # Execute in the module's namespace
    exec(compiled, module.__dict__)  # pylint: disable=exec-used


if __name__ == "__main__":
    try:
        # Get the encoded user code
        user_code = base64.b64decode("{encoded_code}").decode()

        # Run the code
        run_user_code(user_code)
    except Exception:
        # Print only the relevant parts of the traceback
        exc_type, exc_value, exc_tb = sys.exc_info()

        # Filter traceback to only show user code frames
        filtered_tb = []
        tb = exc_tb
        while tb is not None:
            if tb.tb_frame.f_code.co_filename == "<target_code>":
                filtered_tb.append(tb)
            tb = tb.tb_next

        if filtered_tb:
            # Create a new traceback from the filtered frames
            exc_value.__traceback__ = filtered_tb[-1]
            traceback.print_exception(exc_type, exc_value, exc_value.__traceback__)
        else:
            # Fallback if no user code frames found - raise the original exception type
            # with the original message but create a fresh traceback
            raise exc_type(str(exc_value)) from None

        sys.exit(1)
"#;
