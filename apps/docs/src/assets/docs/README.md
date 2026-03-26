### How to update SVG diagrams

1. Open XML file (e.g. `src/assets/docs/sandbox-states.drawio.xml`) in [draw.io](https://app.diagrams.net/)
2. Make necessary changes
3. Export as SVG with "Transparent Backround" toggled `on` and "Embed Fonts" toggled `off`
4. Replace the existing SVG with the newly exported one
5. In the new SVG, remove the `color-scheme: light dark;` CSS property declaration
6. Export as XML and replace the existing XML with the newly exported one
