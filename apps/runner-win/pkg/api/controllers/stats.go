// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/daytonaio/runner-win/pkg/libvirt"
	"github.com/daytonaio/runner-win/pkg/runner"
	"github.com/gin-gonic/gin"
)

// MemoryStatsResponse represents the API response for memory stats
type MemoryStatsResponse struct {
	VMNames  []string                    `json:"vm_names"`
	Stats    []libvirt.MemoryStatsRecord `json:"stats"`
	FromTime time.Time                   `json:"from_time"`
	ToTime   time.Time                   `json:"to_time"`
	Count    int                         `json:"count"`
}

// GetMemoryStats returns memory statistics as JSON
//
//	@Summary		Get memory statistics
//	@Description	Returns memory statistics for VMs over a time range
//	@Tags			stats
//	@Produce		json
//	@Param			vm		query		string	false	"Filter by VM name"
//	@Param			hours	query		int		false	"Lookback hours (default: 24)"
//	@Success		200		{object}	MemoryStatsResponse
//	@Failure		500		{object}	string
//	@Router			/stats/memory [get]
//
//	@id				GetMemoryStats
func GetMemoryStats(ctx *gin.Context) {
	r := runner.GetInstance(nil)
	if r.StatsStore == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Stats store not initialized"})
		return
	}

	// Parse query parameters
	vmName := ctx.Query("vm")
	hoursStr := ctx.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	// Calculate time range
	toTime := time.Now()
	fromTime := toTime.Add(-time.Duration(hours) * time.Hour)

	// Get stats
	stats, err := r.StatsStore.GetMemoryStats(ctx.Request.Context(), vmName, fromTime, toTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get VM names
	vmNames, err := r.StatsStore.GetAllVMNames(ctx.Request.Context())
	if err != nil {
		vmNames = []string{}
	}

	response := MemoryStatsResponse{
		VMNames:  vmNames,
		Stats:    stats,
		FromTime: fromTime,
		ToTime:   toTime,
		Count:    len(stats),
	}

	ctx.JSON(http.StatusOK, response)
}

// GetMemoryStatsView returns an HTML visualization page for memory stats
//
//	@Summary		Memory stats visualization
//	@Description	Returns an interactive HTML page with memory usage charts
//	@Tags			stats
//	@Produce		html
//	@Param			hours	query		int		false	"Lookback hours (default: 24)"
//	@Success		200		{string}	string	"HTML page"
//	@Router			/stats/memory/view [get]
//
//	@id				GetMemoryStatsView
func GetMemoryStatsView(ctx *gin.Context) {
	hoursStr := ctx.DefaultQuery("hours", "24")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Memory Stats - Daytona Runner</title>
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
            background: linear-gradient(90deg, #00d4ff, #7b2ff7);
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
            border-color: #00d4ff;
        }
        
        button.primary {
            background: linear-gradient(135deg, #00d4ff, #7b2ff7);
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
            color: #00d4ff;
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
            border-top-color: #00d4ff;
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
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Memory Stats Dashboard</h1>
            <div class="controls">
                <select id="hours-select">
                    <option value="1">Last 1 hour</option>
                    <option value="6">Last 6 hours</option>
                    <option value="24" selected>Last 24 hours</option>
                    <option value="48">Last 48 hours</option>
                    <option value="168">Last 7 days</option>
                </select>
                <select id="vm-filter">
                    <option value="">All VMs</option>
                </select>
                <button class="primary" onclick="refreshData()">Refresh</button>
                <span class="refresh-indicator" id="last-refresh"></span>
            </div>
        </header>
        
        <div class="stats-grid" id="stats-grid">
            <div class="stat-card">
                <div class="label">Total VMs</div>
                <div class="value" id="total-vms">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Max Memory</div>
                <div class="value" id="total-max-memory">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Actual Allocated</div>
                <div class="value warning" id="total-actual-memory">-</div>
            </div>
            <div class="stat-card">
                <div class="label">Memory Saved</div>
                <div class="value success" id="memory-saved">-</div>
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
                <span class="chart-title">Latest VM Status</span>
            </div>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>VM Name</th>
                            <th>Max Memory</th>
                            <th>Actual</th>
                            <th>Used</th>
                            <th>Unused</th>
                            <th>Balloon</th>
                            <th>Last Updated</th>
                        </tr>
                    </thead>
                    <tbody id="vm-table-body">
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    
    <script>
        let chart = null;
        let allData = null;
        let vmColors = {};
        let disabledVMs = new Set();
        
        const colorPalette = [
            '#00d4ff', '#7b2ff7', '#50fa7b', '#ffb347', '#ff6b6b',
            '#4ecdc4', '#f7dc6f', '#bb8fce', '#85c1e9', '#f8b500'
        ];
        
        function getVMColor(vmName, index) {
            if (!vmColors[vmName]) {
                vmColors[vmName] = colorPalette[index % colorPalette.length];
            }
            return vmColors[vmName];
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
        
        async function fetchData() {
            const hours = document.getElementById('hours-select').value;
            const vm = document.getElementById('vm-filter').value;
            
            let url = '/stats/memory?hours=' + hours;
            if (vm) url += '&vm=' + encodeURIComponent(vm);
            
            const response = await fetch(url);
            return await response.json();
        }
        
        function updateStats(data) {
            document.getElementById('data-points').textContent = data.count;
            
            // Calculate totals from latest stats per VM
            const latestByVM = {};
            data.stats.forEach(s => {
                if (!latestByVM[s.vm_name] || new Date(s.timestamp) > new Date(latestByVM[s.vm_name].timestamp)) {
                    latestByVM[s.vm_name] = s;
                }
            });
            
            const vms = Object.values(latestByVM);
            const totalMax = vms.reduce((sum, v) => sum + v.max_memory_kib, 0);
            const totalActual = vms.reduce((sum, v) => sum + v.actual_kib, 0);
            const saved = totalMax - totalActual;
            
            document.getElementById('total-vms').textContent = vms.length;
            document.getElementById('total-max-memory').textContent = formatMemory(totalMax);
            document.getElementById('total-actual-memory').textContent = formatMemory(totalActual);
            document.getElementById('memory-saved').textContent = formatMemory(saved);
            
            // Update VM filter dropdown
            const vmFilter = document.getElementById('vm-filter');
            const currentValue = vmFilter.value;
            vmFilter.innerHTML = '<option value="">All VMs</option>';
            data.vm_names.forEach(name => {
                const opt = document.createElement('option');
                opt.value = name;
                opt.textContent = name;
                if (name === currentValue) opt.selected = true;
                vmFilter.appendChild(opt);
            });
            
            // Update table
            const tbody = document.getElementById('vm-table-body');
            tbody.innerHTML = '';
            vms.forEach(v => {
                const tr = document.createElement('tr');
                tr.innerHTML = ` + "`" + `
                    <td>${v.vm_name}</td>
                    <td>${formatMemory(v.max_memory_kib)}</td>
                    <td>${formatMemory(v.actual_kib)}</td>
                    <td>${formatMemory(v.used_kib)}</td>
                    <td>${formatMemory(v.unused_kib)}</td>
                    <td><span class="status-badge ${v.balloon_active ? 'status-active' : 'status-inactive'}">${v.balloon_active ? 'Active' : 'Inactive'}</span></td>
                    <td>${formatTime(v.timestamp)}</td>
                ` + "`" + `;
                tbody.appendChild(tr);
            });
        }
        
        function updateChart(data) {
            const ctx = document.getElementById('memory-chart').getContext('2d');
            
            // Group data by VM
            const vmData = {};
            data.stats.forEach(s => {
                if (!vmData[s.vm_name]) {
                    vmData[s.vm_name] = [];
                }
                vmData[s.vm_name].push({
                    x: new Date(s.timestamp),
                    actual: s.actual_kib / (1024 * 1024), // Convert to GB
                    max: s.max_memory_kib / (1024 * 1024),
                    used: s.used_kib / (1024 * 1024),
                    unused: s.unused_kib / (1024 * 1024)
                });
            });
            
            // Create datasets
            const datasets = [];
            let colorIndex = 0;
            
            Object.entries(vmData).forEach(([vmName, points]) => {
                const color = getVMColor(vmName, colorIndex++);
                const isDisabled = disabledVMs.has(vmName);
                
                // Actual memory line
                datasets.push({
                    label: vmName + ' (Actual)',
                    data: points.map(p => ({ x: p.x, y: p.actual })),
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
                    label: vmName + ' (Max)',
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
            updateLegend(Object.keys(vmData));
        }
        
        function updateLegend(vmNames) {
            const legend = document.getElementById('vm-legend');
            legend.innerHTML = '';
            
            vmNames.forEach((vmName, index) => {
                const color = getVMColor(vmName, index);
                const isDisabled = disabledVMs.has(vmName);
                
                const item = document.createElement('div');
                item.className = 'vm-legend-item' + (isDisabled ? ' disabled' : '');
                item.innerHTML = ` + "`" + `
                    <div class="vm-legend-color" style="background: ${color}"></div>
                    <span class="vm-legend-name">${vmName}</span>
                ` + "`" + `;
                item.onclick = () => toggleVM(vmName);
                legend.appendChild(item);
            });
        }
        
        function toggleVM(vmName) {
            if (disabledVMs.has(vmName)) {
                disabledVMs.delete(vmName);
            } else {
                disabledVMs.add(vmName);
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
        document.getElementById('vm-filter').addEventListener('change', refreshData);
    </script>
</body>
</html>`

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}
