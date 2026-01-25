// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/gin-gonic/gin"
)

var statsStore *cloudhypervisor.StatsStore

// SetStatsStore sets the stats store instance for the controllers
func SetStatsStore(store *cloudhypervisor.StatsStore) {
	statsStore = store
}

// MemoryStatsResponse represents the API response for memory stats
type MemoryStatsResponse struct {
	SandboxIds []string                            `json:"sandbox_ids"`
	Stats      []cloudhypervisor.MemoryStatsRecord `json:"stats"`
	FromTime   time.Time                           `json:"from_time"`
	ToTime     time.Time                           `json:"to_time"`
	Count      int                                 `json:"count"`
}

// GetMemoryStatsJSON returns memory statistics as JSON
//
//	@Summary		Get memory statistics
//	@Description	Returns memory statistics for sandboxes over a time range
//	@Tags			stats
//	@Produce		json
//	@Param			sandbox	query		string	false	"Filter by sandbox ID"
//	@Param			hours	query		int		false	"Lookback hours (default: 24)"
//	@Success		200		{object}	MemoryStatsResponse
//	@Failure		500		{object}	string
//	@Router			/stats/memory [get]
//
//	@id				GetMemoryStatsJSON
func GetMemoryStatsJSON(ctx *gin.Context) {
	if statsStore == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Stats store not initialized"})
		return
	}

	// Parse query parameters
	sandboxId := ctx.Query("sandbox")
	hoursStr := ctx.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	// Calculate time range
	toTime := time.Now()
	fromTime := toTime.Add(-time.Duration(hours) * time.Hour)

	// Get stats
	stats, err := statsStore.GetMemoryStats(ctx.Request.Context(), sandboxId, fromTime, toTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get sandbox IDs
	sandboxIds, err := statsStore.GetAllSandboxIds(ctx.Request.Context())
	if err != nil {
		sandboxIds = []string{}
	}

	response := MemoryStatsResponse{
		SandboxIds: sandboxIds,
		Stats:      stats,
		FromTime:   fromTime,
		ToTime:     toTime,
		Count:      len(stats),
	}

	ctx.JSON(http.StatusOK, response)
}

// GetMemoryStatsViewHTML returns an HTML visualization page for memory stats
//
//	@Summary		Memory stats visualization
//	@Description	Returns an interactive HTML page with memory usage charts
//	@Tags			stats
//	@Produce		html
//	@Param			hours	query		int		false	"Lookback hours (default: 24)"
//	@Success		200		{string}	string	"HTML page"
//	@Router			/stats/memory/view [get]
//
//	@id				GetMemoryStatsViewHTML
func GetMemoryStatsViewHTML(ctx *gin.Context) {
	hoursStr := ctx.DefaultQuery("hours", "24")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Memory Stats - Cloud Hypervisor Runner</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@3.0.0/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'SF Mono', 'Fira Code', 'JetBrains Mono', monospace;
            background: linear-gradient(135deg, #0f0f23 0%, #1a1a2e 50%, #16213e 100%);
            color: #e0e0e0;
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 1600px;
            margin: 0 auto;
        }
        
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        h1 {
            font-size: 1.8rem;
            font-weight: 600;
            background: linear-gradient(90deg, #ff6b35, #f7931e);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .controls {
            display: flex;
            gap: 15px;
            align-items: center;
        }
        
        select, button {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            color: #e0e0e0;
            padding: 10px 16px;
            border-radius: 8px;
            font-family: inherit;
            font-size: 0.9rem;
            cursor: pointer;
            transition: all 0.2s ease;
        }
        
        select:hover, button:hover {
            background: rgba(255, 255, 255, 0.1);
            border-color: #ff6b35;
        }
        
        button.primary {
            background: linear-gradient(135deg, #ff6b35, #f7931e);
            border: none;
            color: white;
            font-weight: 500;
        }
        
        button.primary:hover {
            opacity: 0.9;
            transform: translateY(-1px);
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.08);
            border-radius: 12px;
            padding: 20px;
            text-align: center;
        }
        
        .stat-card .label {
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #888;
            margin-bottom: 8px;
        }
        
        .stat-card .value {
            font-size: 1.8rem;
            font-weight: 700;
            color: #ff6b35;
        }
        
        .stat-card .value.warning {
            color: #ffb347;
        }
        
        .stat-card .value.success {
            color: #50fa7b;
        }
        
        .chart-container {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.08);
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 30px;
        }
        
        .chart-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
        }
        
        .chart-title {
            font-size: 1.1rem;
            font-weight: 500;
        }
        
        .chart-wrapper {
            position: relative;
            height: 400px;
        }
        
        .vm-legend {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid rgba(255, 255, 255, 0.05);
        }
        
        .vm-legend-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 6px 12px;
            background: rgba(255, 255, 255, 0.03);
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.2s;
        }
        
        .vm-legend-item:hover {
            background: rgba(255, 255, 255, 0.08);
        }
        
        .vm-legend-item.disabled {
            opacity: 0.4;
        }
        
        .vm-legend-color {
            width: 12px;
            height: 12px;
            border-radius: 3px;
        }
        
        .vm-legend-name {
            font-size: 0.85rem;
            font-family: monospace;
        }
        
        .loading {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 300px;
            color: #888;
        }
        
        .loading::after {
            content: '';
            width: 30px;
            height: 30px;
            border: 3px solid rgba(255, 255, 255, 0.1);
            border-top-color: #ff6b35;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-left: 15px;
        }
        
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
        
        .refresh-indicator {
            font-size: 0.8rem;
            color: #666;
        }
        
        .no-data {
            text-align: center;
            padding: 60px;
            color: #666;
        }
        
        .table-container {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.08);
            border-radius: 12px;
            overflow: hidden;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
        }
        
        th, td {
            padding: 12px 16px;
            text-align: left;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }
        
        th {
            background: rgba(255, 255, 255, 0.03);
            font-weight: 500;
            font-size: 0.8rem;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: #888;
        }
        
        tr:hover {
            background: rgba(255, 255, 255, 0.02);
        }
        
        .status-badge {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 20px;
            font-size: 0.75rem;
            font-weight: 500;
        }
        
        .status-active {
            background: rgba(80, 250, 123, 0.15);
            color: #50fa7b;
        }
        
        .status-inactive {
            background: rgba(255, 85, 85, 0.15);
            color: #ff5555;
        }
        
        .sandbox-id {
            font-family: monospace;
            font-size: 0.85rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔥 Cloud Hypervisor Memory Stats</h1>
            <div class="controls">
                <select id="hours-select">
                    <option value="1">Last 1 hour</option>
                    <option value="6">Last 6 hours</option>
                    <option value="24" selected>Last 24 hours</option>
                    <option value="48">Last 48 hours</option>
                    <option value="168">Last 7 days</option>
                </select>
                <select id="sandbox-filter">
                    <option value="">All Sandboxes</option>
                </select>
                <button class="primary" onclick="refreshData()">Refresh</button>
                <span class="refresh-indicator" id="last-refresh"></span>
            </div>
        </header>
        
        <div class="stats-grid" id="stats-grid">
            <div class="stat-card">
                <div class="label">Total Sandboxes</div>
                <div class="value" id="total-sandboxes">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Max Memory</div>
                <div class="value" id="total-max-memory">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Available Memory</div>
                <div class="value warning" id="total-available-memory">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Balloon Reclaimed</div>
                <div class="value success" id="balloon-reclaimed">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Data Points</div>
                <div class="value" id="data-points">-</div>
            </div>
        </div>
        
        <div class="chart-container">
            <div class="chart-header">
                <span class="chart-title">Memory Usage Over Time</span>
            </div>
            <div class="chart-wrapper">
                <canvas id="memory-chart"></canvas>
            </div>
            <div class="vm-legend" id="vm-legend"></div>
        </div>
        
        <div class="chart-container">
            <div class="chart-header">
                <span class="chart-title">Latest Sandbox Status</span>
            </div>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>Sandbox ID</th>
                            <th>Max Memory</th>
                            <th>Available</th>
                            <th>Used</th>
                            <th>Balloon Size</th>
                            <th>Status</th>
                            <th>Last Updated</th>
                        </tr>
                    </thead>
                    <tbody id="sandbox-table-body">
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    
    <script>
        let chart = null;
        let allData = null;
        let sandboxColors = {};
        let disabledSandboxes = new Set();
        
        const colorPalette = [
            '#ff6b35', '#f7931e', '#50fa7b', '#00d4ff', '#7b2ff7',
            '#4ecdc4', '#f7dc6f', '#bb8fce', '#85c1e9', '#f8b500'
        ];
        
        function getSandboxColor(sandboxId, index) {
            if (!sandboxColors[sandboxId]) {
                sandboxColors[sandboxId] = colorPalette[index % colorPalette.length];
            }
            return sandboxColors[sandboxId];
        }
        
        function formatMemory(kib) {
            if (kib >= 1024 * 1024) {
                return (kib / (1024 * 1024)).toFixed(1) + ' GB';
            }
            if (kib >= 1024) {
                return (kib / 1024).toFixed(1) + ' MB';
            }
            return kib + ' KiB';
        }
        
        function formatTime(timestamp) {
            return new Date(timestamp).toLocaleString();
        }
        
        function shortId(id) {
            return id.substring(0, 8) + '...';
        }
        
        async function fetchData() {
            const hours = document.getElementById('hours-select').value;
            const sandbox = document.getElementById('sandbox-filter').value;
            
            let url = '/stats/memory?hours=' + hours;
            if (sandbox) url += '&sandbox=' + encodeURIComponent(sandbox);
            
            const response = await fetch(url);
            return await response.json();
        }
        
        function updateStats(data) {
            document.getElementById('data-points').textContent = data.count;
            
            // Calculate totals from latest stats per sandbox
            const latestBySandbox = {};
            data.stats.forEach(s => {
                if (!latestBySandbox[s.sandbox_id] || new Date(s.timestamp) > new Date(latestBySandbox[s.sandbox_id].timestamp)) {
                    latestBySandbox[s.sandbox_id] = s;
                }
            });
            
            const sandboxes = Object.values(latestBySandbox);
            const totalMax = sandboxes.reduce((sum, v) => sum + v.max_memory_kib, 0);
            const totalAvailable = sandboxes.reduce((sum, v) => sum + v.available_kib, 0);
            const totalBalloon = sandboxes.reduce((sum, v) => sum + v.balloon_size_kib, 0);
            
            document.getElementById('total-sandboxes').textContent = sandboxes.length;
            document.getElementById('total-max-memory').textContent = formatMemory(totalMax);
            document.getElementById('total-available-memory').textContent = formatMemory(totalAvailable);
            document.getElementById('balloon-reclaimed').textContent = formatMemory(totalBalloon);
            
            // Update sandbox filter dropdown
            const sandboxFilter = document.getElementById('sandbox-filter');
            const currentValue = sandboxFilter.value;
            sandboxFilter.innerHTML = '<option value="">All Sandboxes</option>';
            data.sandbox_ids.forEach(id => {
                const opt = document.createElement('option');
                opt.value = id;
                opt.textContent = shortId(id);
                if (id === currentValue) opt.selected = true;
                sandboxFilter.appendChild(opt);
            });
            
            // Update table
            const tbody = document.getElementById('sandbox-table-body');
            tbody.innerHTML = '';
            sandboxes.forEach(v => {
                const tr = document.createElement('tr');
                tr.innerHTML = ` + "`" + `
                    <td class="sandbox-id" title="${v.sandbox_id}">${shortId(v.sandbox_id)}</td>
                    <td>${formatMemory(v.max_memory_kib)}</td>
                    <td>${formatMemory(v.available_kib)}</td>
                    <td>${formatMemory(v.used_kib)}</td>
                    <td>${formatMemory(v.balloon_size_kib)}</td>
                    <td><span class="status-badge ${v.balloon_active ? 'status-active' : 'status-inactive'}">${v.balloon_active ? 'Active' : 'Inactive'}</span></td>
                    <td>${formatTime(v.timestamp)}</td>
                ` + "`" + `;
                tbody.appendChild(tr);
            });
        }
        
        function updateChart(data) {
            const ctx = document.getElementById('memory-chart').getContext('2d');
            
            // Group data by sandbox
            const sandboxData = {};
            data.stats.forEach(s => {
                if (!sandboxData[s.sandbox_id]) {
                    sandboxData[s.sandbox_id] = [];
                }
                sandboxData[s.sandbox_id].push({
                    x: new Date(s.timestamp),
                    available: s.available_kib / (1024 * 1024), // Convert to GB
                    max: s.max_memory_kib / (1024 * 1024),
                    used: s.used_kib / (1024 * 1024),
                    balloon: s.balloon_size_kib / (1024 * 1024)
                });
            });
            
            // Create datasets
            const datasets = [];
            let colorIndex = 0;
            
            Object.entries(sandboxData).forEach(([sandboxId, points]) => {
                const color = getSandboxColor(sandboxId, colorIndex++);
                const isDisabled = disabledSandboxes.has(sandboxId);
                const shortName = sandboxId.substring(0, 8);
                
                // Used memory line
                datasets.push({
                    label: shortName + ' (Used)',
                    data: points.map(p => ({ x: p.x, y: p.used })),
                    borderColor: color,
                    backgroundColor: color + '20',
                    fill: true,
                    tension: 0.3,
                    borderWidth: 2,
                    pointRadius: 0,
                    hidden: isDisabled
                });
                
                // Max memory line (dashed)
                datasets.push({
                    label: shortName + ' (Max)',
                    data: points.map(p => ({ x: p.x, y: p.max })),
                    borderColor: color,
                    borderDash: [5, 5],
                    fill: false,
                    tension: 0,
                    borderWidth: 1,
                    pointRadius: 0,
                    hidden: isDisabled
                });
            });
            
            if (chart) {
                chart.data.datasets = datasets;
                chart.update('none');
            } else {
                chart = new Chart(ctx, {
                    type: 'line',
                    data: { datasets },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        interaction: {
                            mode: 'index',
                            intersect: false
                        },
                        plugins: {
                            legend: {
                                display: false
                            },
                            tooltip: {
                                backgroundColor: 'rgba(15, 15, 35, 0.95)',
                                titleColor: '#fff',
                                bodyColor: '#ccc',
                                borderColor: 'rgba(255, 255, 255, 0.1)',
                                borderWidth: 1,
                                padding: 12,
                                callbacks: {
                                    label: function(context) {
                                        return context.dataset.label + ': ' + context.parsed.y.toFixed(2) + ' GB';
                                    }
                                }
                            }
                        },
                        scales: {
                            x: {
                                type: 'time',
                                time: {
                                    displayFormats: {
                                        hour: 'HH:mm',
                                        day: 'MMM d'
                                    }
                                },
                                grid: {
                                    color: 'rgba(255, 255, 255, 0.05)'
                                },
                                ticks: {
                                    color: '#888'
                                }
                            },
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'Memory (GB)',
                                    color: '#888'
                                },
                                grid: {
                                    color: 'rgba(255, 255, 255, 0.05)'
                                },
                                ticks: {
                                    color: '#888'
                                }
                            }
                        }
                    }
                });
            }
            
            // Update legend
            updateLegend(Object.keys(sandboxData));
        }
        
        function updateLegend(sandboxIds) {
            const legend = document.getElementById('vm-legend');
            legend.innerHTML = '';
            
            sandboxIds.forEach((sandboxId, index) => {
                const color = getSandboxColor(sandboxId, index);
                const isDisabled = disabledSandboxes.has(sandboxId);
                const shortName = sandboxId.substring(0, 8);
                
                const item = document.createElement('div');
                item.className = 'vm-legend-item' + (isDisabled ? ' disabled' : '');
                item.title = sandboxId;
                item.innerHTML = ` + "`" + `
                    <div class="vm-legend-color" style="background: ${color}"></div>
                    <span class="vm-legend-name">${shortName}</span>
                ` + "`" + `;
                item.onclick = () => toggleSandbox(sandboxId);
                legend.appendChild(item);
            });
        }
        
        function toggleSandbox(sandboxId) {
            if (disabledSandboxes.has(sandboxId)) {
                disabledSandboxes.delete(sandboxId);
            } else {
                disabledSandboxes.add(sandboxId);
            }
            updateChart(allData);
        }
        
        async function refreshData() {
            try {
                allData = await fetchData();
                updateStats(allData);
                updateChart(allData);
                document.getElementById('last-refresh').textContent = 'Updated: ' + new Date().toLocaleTimeString();
            } catch (err) {
                console.error('Failed to fetch data:', err);
            }
        }
        
        // Initial load
        document.getElementById('hours-select').value = '` + hoursStr + `';
        refreshData();
        
        // Auto-refresh every 30 seconds
        setInterval(refreshData, 30000);
        
        // Refresh on filter change
        document.getElementById('hours-select').addEventListener('change', refreshData);
        document.getElementById('sandbox-filter').addEventListener('change', refreshData);
    </script>
</body>
</html>`

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}
