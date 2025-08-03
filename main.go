package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"hostscanner/network"
	"hostscanner/scanner"
)

// HostScannerUI represents the terminal user interface for the host scanner.
type HostScannerUI struct {
	app          *tview.Application
	pages        *tview.Pages
	mainLayout   *tview.Flex
	sidebar      *tview.Flex
	contentArea  *tview.Flex
	header       *tview.TextView
	footer       *tview.TextView
	table        *tview.Table
	infoPanel    *tview.TextView
	progressBar  *tview.TextView
	scanButton   *tview.Button
	ipInput      *tview.InputField
	showInactive *tview.Checkbox
	isScanning   bool
	scanResults  *scanner.ScanResult
}

func main() {
	ui := NewHostScannerUI()
	if err := ui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// NewHostScannerUI creates a new instance of the host scanner UI.
func NewHostScannerUI() *HostScannerUI {
	ui := &HostScannerUI{
		app: tview.NewApplication(),
	}

	ui.setupModernUI()
	return ui
}

func (ui *HostScannerUI) setupModernUI() {
	// Set up modern color scheme
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlue
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorGreen
	tview.Styles.BorderColor = tcell.ColorDarkSlateGray
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorGray
	tview.Styles.TertiaryTextColor = tcell.ColorDarkGray
	tview.Styles.InverseTextColor = tcell.ColorBlack

	ui.createHeader()
	ui.createSidebar()
	ui.createContentArea()
	ui.createFooter()
	ui.setupLayout()
	ui.setupPages()
}

func (ui *HostScannerUI) createHeader() {
	ui.header = tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Simple, clean header design
	headerText := `[#00ff88::b]HostScanner [#666666::]• Network Discovery Tool`

	ui.header.SetText(headerText)
}

func (ui *HostScannerUI) createSidebar() {
	// IP Input with modern styling
	ui.ipInput = tview.NewInputField().
		SetLabel("🎯 Target Range ").
		SetText("192.168.1.0/24").
		SetFieldWidth(0).
		SetLabelColor(tcell.ColorLightBlue).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldTextColor(tcell.ColorWhite)

	// Checkbox with modern styling
	ui.showInactive = tview.NewCheckbox().
		SetLabel("👻 Show offline hosts").
		SetLabelColor(tcell.ColorLightGray).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldTextColor(tcell.ColorWhite)

	// Scan button with modern styling
	ui.scanButton = tview.NewButton("🚀 Start Scan")
	ui.scanButton.SetSelectedFunc(ui.scanNetwork).
		SetLabelColor(tcell.ColorBlack).
		SetBackgroundColor(tcell.ColorLightGreen)

	autoDetectBtn := tview.NewButton("🔍 Auto-detect")
	autoDetectBtn.SetSelectedFunc(ui.autoDetectNetwork).
		SetLabelColor(tcell.ColorBlack).
		SetBackgroundColor(tcell.ColorLightBlue)

	quitBtn := tview.NewButton("❌ Quit")
	quitBtn.SetSelectedFunc(func() { ui.app.Stop() }).
		SetLabelColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorDarkRed)

	// Progress bar
	ui.progressBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#444444]Ready to scan")

	// Info panel
	ui.infoPanel = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	ui.infoPanel.SetBorder(true).
		SetBorderColor(tcell.ColorDarkSlateGray).
		SetTitle(" 📊 Scan Statistics ").
		SetTitleColor(tcell.ColorLightBlue)

	ui.updateInfoPanel()

	// Sidebar layout
	ui.sidebar = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("[#00ff88::b]⚙️  Configuration").SetDynamicColors(true), 1, 0, false).
		AddItem(tview.NewTextView(), 1, 0, false). // Spacer
		AddItem(ui.ipInput, 1, 0, false).
		AddItem(tview.NewTextView(), 1, 0, false). // Spacer
		AddItem(ui.showInactive, 1, 0, false).
		AddItem(tview.NewTextView(), 2, 0, false). // Spacer
		AddItem(ui.scanButton, 1, 0, false).
		AddItem(tview.NewTextView(), 1, 0, false). // Spacer
		AddItem(autoDetectBtn, 1, 0, false).
		AddItem(tview.NewTextView(), 1, 0, false). // Spacer
		AddItem(quitBtn, 1, 0, false).
		AddItem(tview.NewTextView(), 2, 0, false). // Spacer
		AddItem(ui.progressBar, 1, 0, false).
		AddItem(tview.NewTextView(), 1, 0, false). // Spacer
		AddItem(ui.infoPanel, 0, 1, false)

	ui.sidebar.SetBorder(true).
		SetBorderColor(tcell.ColorDarkSlateGray).
		SetTitle(" 🎛️  Control Panel ").
		SetTitleColor(tcell.ColorLightGreen)
}

func (ui *HostScannerUI) createContentArea() {
	ui.table = tview.NewTable().
		SetBorders(false).
		SetSeparator('│').
		SetSelectable(true, false).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorWhite)).
		SetFixed(1, 0)

	ui.setupModernTable()

	ui.contentArea = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.table, 0, 1, false)

	ui.contentArea.SetBorder(true).
		SetBorderColor(tcell.ColorDarkSlateGray).
		SetTitle(" 📋 Network Devices ").
		SetTitleColor(tcell.ColorLightCyan)
}

func (ui *HostScannerUI) setupModernTable() {
	// Modern table headers with icons and styling
	headers := []struct {
		text  string
		align int
	}{
		{"🔗 Status", tview.AlignCenter},
		{"🌐 IP Address", tview.AlignLeft},
		{"🏠 Hostname", tview.AlignLeft},
		{"🔧 MAC Address", tview.AlignLeft},
		{"🏢 Vendor", tview.AlignLeft},
		{"⚡ Latency", tview.AlignRight},
	}

	// Define expansion settings for each column to match data cells
	expansions := []int{0, 0, 1, 0, 1, 0} // Status, IP, Hostname, MAC, Vendor, Latency

	for col, header := range headers {
		cell := tview.NewTableCell(header.text).
			SetAlign(header.align).
			SetSelectable(false).
			SetBackgroundColor(tcell.ColorDarkSlateGray).
			SetTextColor(tcell.ColorLightCyan).
			SetAttributes(tcell.AttrBold).
			SetExpansion(expansions[col])
		ui.table.SetCell(0, col, cell)
	}
}

func (ui *HostScannerUI) createFooter() {
	ui.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#444444]Press [#00ff88::b]Tab[#444444] to navigate • [#00ff88::b]Enter[#444444] to select • [#00ff88::b]Ctrl+C[#444444] to quit")
}

func (ui *HostScannerUI) setupLayout() {
	ui.mainLayout = tview.NewFlex().
		AddItem(ui.sidebar, 0, 1, true).
		AddItem(ui.contentArea, 0, 3, false)

	mainWithFooter := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.header, 1, 0, false).
		AddItem(ui.mainLayout, 0, 1, false).
		AddItem(ui.footer, 1, 0, false)

	ui.pages = tview.NewPages().
		AddPage("main", mainWithFooter, true, true)
}

func (ui *HostScannerUI) setupPages() {
	ui.app.SetRoot(ui.pages, true).EnableMouse(true)
}

func (ui *HostScannerUI) updateInfoPanel() {
	if ui.scanResults == nil {
		ui.infoPanel.SetText(`[#888888]No scan data available

[#00ff88::b]💡 Quick Start:
[#ffffff]1. Enter IP range
[#ffffff]2. Click Start Scan
[#ffffff]3. View results

[#ffaa00::b]📝 Supported formats:
[#ffffff]• CIDR: 192.168.1.0/24
[#ffffff]• Range: 192.168.1.1-100
[#ffffff]• Single: 192.168.1.1`)
		return
	}

	activeHosts := ui.scanResults.AliveHosts
	totalHosts := ui.scanResults.TotalHosts
	scanTime := ui.scanResults.ScanTime

	percentage := float64(activeHosts) / float64(totalHosts) * 100

	statusColor := "#ff4444"
	if percentage > 50 {
		statusColor = "#ffaa00"
	}
	if percentage > 80 {
		statusColor = "#00ff88"
	}

	info := fmt.Sprintf(`[#00ff88::b]📊 Scan Summary

[#ffffff::b]Total Hosts:[#ffffff] %d
[#00ff88::b]Active Hosts:[#ffffff] %d
[%s::b]Success Rate:[#ffffff] %.1f%%

[#ffffff::b]Scan Duration:[#ffffff] %v

[#888888]Last updated: %s`,
		totalHosts,
		activeHosts,
		statusColor, percentage,
		scanTime.Truncate(time.Millisecond),
		time.Now().Format("15:04:05"))

	ui.infoPanel.SetText(info)
}

func (ui *HostScannerUI) Run() error {
	return ui.app.Run()
}

func (ui *HostScannerUI) scanNetwork() {
	if ui.isScanning {
		return
	}

	ipRange := ui.ipInput.GetText()
	if ipRange == "" {
		ui.showModernError("Please enter an IP range")
		return
	}

	ui.isScanning = true
	ui.scanButton.SetLabel("⏳ Scanning...")
	ui.scanButton.SetBackgroundColor(tcell.ColorOrange)
	ui.updateProgressBar("Initializing scan...", 0)

	// Parse IP range
	ipr, err := network.ParseIPRange(ipRange)
	if err != nil {
		ui.showModernError(fmt.Sprintf("Invalid IP range: %v", err))
		ui.resetScanButton()
		return
	}

	// Clear previous results
	ui.clearTable()

	// Start scanning in goroutine
	go func() {
		ips := ipr.GenerateIPs()
		ui.app.QueueUpdateDraw(func() {
			ui.updateProgressBar(fmt.Sprintf("Scanning %d hosts...", len(ips)), 25)
		})

		result := scanner.ScanNetwork(ips, time.Second, 100)

		ui.app.QueueUpdateDraw(func() {
			ui.scanResults = result
			ui.displayModernResults(result, ipRange)
			ui.updateInfoPanel()
			ui.resetScanButton()
			ui.updateProgressBar("Scan completed!", 100)
		})
	}()
}

func (ui *HostScannerUI) autoDetectNetwork() {
	localNetwork, err := network.GetLocalNetworkRange()
	if err != nil {
		ui.showModernError(fmt.Sprintf("Failed to detect local network: %v", err))
		return
	}

	ui.ipInput.SetText(localNetwork)
	ui.updateProgressBar(fmt.Sprintf("Auto-detected: %s", localNetwork), 0)
}

func (ui *HostScannerUI) resetScanButton() {
	ui.isScanning = false
	ui.scanButton.SetLabel("🚀 Start Scan")
	ui.scanButton.SetBackgroundColor(tcell.ColorLightGreen)
}

func (ui *HostScannerUI) updateProgressBar(message string, progress int) {
	var progressBar strings.Builder
	barWidth := 20
	filled := int(float64(barWidth) * float64(progress) / 100.0)

	progressBar.WriteString("[#00ff88]")
	for i := 0; i < filled; i++ {
		progressBar.WriteString("█")
	}
	progressBar.WriteString("[#333333]")
	for i := filled; i < barWidth; i++ {
		progressBar.WriteString("░")
	}

	ui.progressBar.SetText(fmt.Sprintf("[#ffffff]%s\n%s [#00ff88]%d%%", message, progressBar.String(), progress))
}

func (ui *HostScannerUI) clearTable() {
	ui.table.Clear()
	ui.setupModernTable()
}

func (ui *HostScannerUI) displayModernResults(result *scanner.ScanResult, ipRange string) {
	showInactive := ui.showInactive.IsChecked()

	row := 1
	for _, host := range result.Hosts {
		if !host.IsAlive && !showInactive {
			continue
		}

		// Modern status indicators with colors
		var status string
		var statusColor tcell.Color
		if host.IsAlive {
			status = "🟢 Online"
			statusColor = tcell.ColorGreen
		} else {
			status = "🔴 Offline"
			statusColor = tcell.ColorRed
		}

		hostname := host.Hostname
		if hostname == "" {
			hostname = "[#666666]Unknown"
		}

		mac := host.MAC
		if mac == "" {
			mac = "[#666666]Unknown"
		}

		vendor := host.Vendor
		if vendor == "" {
			vendor = "[#666666]Unknown"
		}

		latency := fmt.Sprintf("%.1fms", float64(host.Latency.Nanoseconds())/1000000)
		if !host.IsAlive {
			latency = "[#666666]N/A"
		} else {
			// Color code latency
			latencyMs := float64(host.Latency.Nanoseconds()) / 1000000
			if latencyMs < 10 {
				latency = fmt.Sprintf("[#00ff88]%.1fms", latencyMs)
			} else if latencyMs < 50 {
				latency = fmt.Sprintf("[#ffaa00]%.1fms", latencyMs)
			} else {
				latency = fmt.Sprintf("[#ff4444]%.1fms", latencyMs)
			}
		}

		// Create cells with modern styling and responsive expansion
		ui.table.SetCell(row, 0, tview.NewTableCell(status).
			SetAlign(tview.AlignCenter).
			SetTextColor(statusColor).
			SetExpansion(0))

		ui.table.SetCell(row, 1, tview.NewTableCell(host.IP.String()).
			SetAlign(tview.AlignLeft).
			SetTextColor(tcell.ColorLightBlue).
			SetExpansion(0))

		ui.table.SetCell(row, 2, tview.NewTableCell(hostname).
			SetAlign(tview.AlignLeft).
			SetTextColor(tcell.ColorWhite).
			SetExpansion(1))

		ui.table.SetCell(row, 3, tview.NewTableCell(mac).
			SetAlign(tview.AlignLeft).
			SetTextColor(tcell.ColorLightGray).
			SetExpansion(0))

		ui.table.SetCell(row, 4, tview.NewTableCell(vendor).
			SetAlign(tview.AlignLeft).
			SetTextColor(tcell.ColorLightYellow).
			SetExpansion(1))

		ui.table.SetCell(row, 5, tview.NewTableCell(latency).
			SetAlign(tview.AlignRight).
			SetTextColor(tcell.ColorWhite).
			SetExpansion(0))

		// Alternate row colors for better readability
		if row%2 == 0 {
			for col := 0; col < 6; col++ {
				ui.table.GetCell(row, col).SetBackgroundColor(tcell.ColorDarkSlateGray)
			}
		}

		row++
	}

	// Update content area title with modern styling
	ui.contentArea.SetTitle(fmt.Sprintf(" 📋 Network Devices - %d Active / %d Total ",
		result.AliveHosts, result.TotalHosts))
}

func (ui *HostScannerUI) showModernError(message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("❌ Error\n\n%s", message)).
		AddButtons([]string{"OK"}).
		SetBackgroundColor(tcell.ColorDarkRed).
		SetTextColor(tcell.ColorWhite).
		SetButtonBackgroundColor(tcell.ColorRed).
		SetButtonTextColor(tcell.ColorWhite).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ui.pages.SwitchToPage("main")
		})

	ui.pages.AddPage("error", modal, true, true)
}
