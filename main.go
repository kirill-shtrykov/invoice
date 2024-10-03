//go:generate fyne bundle -o data.go assets/icons/invoice-icon.png
package main

import (
	"fyne.io/fyne/v2/app"
)

func main() {
	app := app.New()
	app.SetIcon(resourceInvoiceIconPng)

	g := newInvoiceGenerator()
	g.CreateUI(app)
	g.window.ShowAndRun()
}
