package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type invoiceGenerator struct {
	config         *Config
	quantEntry     *widget.Entry
	startDateEntry *widget.Entry
	endDateEntry   *widget.Entry
	priceEntry     *widget.Entry
	window         fyne.Window
}

func (g *invoiceGenerator) CreateUI(app fyne.App) {
	const windowHeight = 100
	const windowWidth = 500

	g.quantEntry = widget.NewEntry()
	g.startDateEntry = widget.NewEntry()
	g.endDateEntry = widget.NewEntry()
	g.priceEntry = widget.NewEntry()
	generateBtn := widget.NewButton("Generate", g.generateInvoice)
	g.window = app.NewWindow("Invoice Generator")
	g.window.Resize(fyne.NewSize(windowWidth, windowHeight))
	config, err := LoadConfig()
	if err != nil {
		showCustomErrorDialog(err, g.window)
		return
	}
	g.config = config

	g.quantEntry.Resize(fyne.NewSize(55, 36))
	g.quantEntry.Text = fmt.Sprintf("%.2f", g.config.Quantity)
	g.quantEntry.Refresh()

	g.startDateEntry.Resize(fyne.NewSize(90, 36))
	g.endDateEntry.Resize(fyne.NewSize(90, 36))
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth-1, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	g.startDateEntry.Text = firstOfMonth.Format("02.01.2006")
	g.endDateEntry.Text = lastOfMonth.Format("02.01.2006")
	g.startDateEntry.Refresh()
	g.endDateEntry.Refresh()

	g.priceEntry.Resize(fyne.NewSize(55, 36))
	g.priceEntry.Text = fmt.Sprintf("%.2f", g.config.Price)
	g.priceEntry.Refresh()

	generateBtn.Resize(fyne.NewSize(100, 36))

	content := container.NewWithoutLayout(g.quantEntry, g.startDateEntry, g.endDateEntry, g.priceEntry, generateBtn)

	var lastX float32 = 10

	for _, elem := range content.Objects {
		size := elem.Size()
		x := lastX + 10
		y := windowHeight/2 - size.Height/2 - 5
		elem.Move(fyne.NewPos(x, y))
		lastX = x + size.Width
	}

	g.window.SetContent(content)
}

func showCustomErrorDialog(err error, win fyne.Window) {
	content := widget.NewLabel(fmt.Sprintf("%v", err))
	dialog := dialog.NewCustom("Error", "Close", content, win)
	dialog.SetOnClosed(func() { os.Exit(1) })
	dialog.Resize(fyne.NewSize(300, 150))
	dialog.Show()
}

func (g *invoiceGenerator) generateInvoice() {
	quantity, err := strconv.ParseFloat(g.quantEntry.Text, 64)
	if err != nil {
		showCustomErrorDialog(err, g.window)
		return
	}
	g.config.Quantity = quantity

	price, err := strconv.ParseFloat(g.priceEntry.Text, 64)
	if err != nil {
		showCustomErrorDialog(err, g.window)
	}
	g.config.Price = price

	g.config.Dates = g.startDateEntry.Text + "-" + g.endDateEntry.Text

	file, err := CreatePDF(g.config)
	if err != nil {
		showCustomErrorDialog(err, g.window)
	}
	if err := openFile(file); err != nil {
		showCustomErrorDialog(err, g.window)
	}
	g.window.Close()
}

func openFile(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filename)
	case "darwin": // macOS
		cmd = exec.Command("open", filename)
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func newInvoiceGenerator() *invoiceGenerator {
	return &invoiceGenerator{}
}
