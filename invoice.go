package main

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/go-pdf/fpdf"
	"gopkg.in/yaml.v2"
)

var primaryColor = color{0x2E, 0x59, 0x80}
var secondaryColor = color{0xF0, 0xF8, 0xFF}
var whiteColor = color{0xFF, 0xFF, 0xFF}
var blackColor = color{0, 0, 0}

//go:embed assets/fonts/cantarell-regular.ttf
var fontCantarellRegular []byte

//go:embed assets/fonts/cantarell-bold.ttf
var fontCantarellBold []byte

//go:embed i18n/*.yaml
var i18n embed.FS

//go:embed config.yaml
var configFile []byte

type color struct {
	r, g, b uint8
}

func (c *color) IntTuple() (int, int, int) {
	return int(c.r), int(c.g), int(c.b)
}

type Config struct {
	Provider Provider `yaml:"provider"`
	Company  Company  `yaml:"company"`
	Langs    []string `yaml:"langs"`
	Price    float64  `yaml:"price"`
	Quantity float64  `yaml:"quantity"`
	Dates    string   `yaml:"dates"`
}

type Provider struct {
	Name    string `yaml:"name"`
	ID      string `yaml:"id"`
	Address string `yaml:"address"`
	Bank    string `yaml:"bank"`
	IBAN    string `yaml:"iban"`
	SWIFT   string `yaml:"swift"`
	Phone   string `yaml:"phone"`
	Email   string `yaml:"email"`
}

type Company struct {
	Name       string `yaml:"name"`
	Address    string `yaml:"address"`
	Additional string `yaml:"additional"`
	ContractID string `yaml:"contract-id"`
}

type Translation struct {
	TaxID       string `yaml:"tax-id"`
	Address     string `yaml:"address"`
	Bank        string `yaml:"bank"`
	Invoice     string `yaml:"invoice"`
	InvoiceID   string `yaml:"invoice-id"`
	InvoiceDate string `yaml:"invoice-date"`
	BillTo      string `yaml:"bill-to"`
	Quantity    string `yaml:"quantity"`
	Description string `yaml:"description"`
	Price       string `yaml:"price"`
	Subtotal    string `yaml:"subtotal"`
	Tax         string `yaml:"tax"`
	Shipping    string `yaml:"shipping"`
	Additional  string `yaml:"additional"`
	Agreement   string `yaml:"agreement"`
}

func LoadConfig() (*Config, error) {
	var config Config

	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return &config, err
	}

	return &config, nil
}

func LoadTranslation(lang string) (*Translation, error) {
	var tr Translation

	yamlData, err := i18n.ReadFile(fmt.Sprintf("i18n/%s.yaml", lang))
	if err != nil {
		return &tr, err
	}

	err = yaml.Unmarshal(yamlData, &tr)
	if err != nil {
		return &tr, err
	}

	return &tr, nil
}

func drawTable(pdf *fpdf.Fpdf, header []string, data [][]string) {
	// Header
	pdf.SetTextColor(whiteColor.IntTuple())
	pdf.SetFillColor(primaryColor.IntTuple())
	pdf.SetFont("Cantarell-Bold", "", 10)
	w := []float64{20, 110, 20, 20}

	for i := 0; i < len(header); i++ {
		pdf.CellFormat(w[i], 5, header[i], "", 0, "C", true, 0, "")
	}
	pdf.Ln(5)

	// Body
	pdf.SetTextColor(blackColor.IntTuple())
	pdf.SetFillColor(secondaryColor.IntTuple())
	pdf.SetFont("Cantarell-Regular", "", 7)
	fill := false

	for _, row := range data {
		pdf.CellFormat(w[0], 5, row[0], "", 0, "L", fill, 0, "")
		pdf.CellFormat(w[1], 5, row[1], "", 0, "L", fill, 0, "")
		pdf.CellFormat(w[2], 5, row[2], "", 0, "L", fill, 0, "")
		pdf.CellFormat(w[3], 5, row[3], "", 1, "L", fill, 0, "")
		fill = !fill
	}
}

func generatePage(pdf *fpdf.Fpdf, config *Config, translation *Translation) {
	const topMargin float64 = 20
	const leftMargin float64 = 20
	const cellMargin float64 = 5
	const pageWidth = 170
	const providerBlockWidth = 100
	const invoiceBlockWidth = 70
	const additionalBlockWidth = 120
	const lineHeight = 5

	pdf.SetMargins(leftMargin, topMargin, -1)
	pdf.AddPage()
	pdf.SetCellMargin(cellMargin)

	// Provider header block
	pdf.SetFont("Cantarell-Bold", "", 16)
	pdf.SetTextColor(primaryColor.IntTuple())
	pdf.SetFillColor(secondaryColor.IntTuple())
	pdf.CellFormat(providerBlockWidth, 10, strings.ToUpper(config.Provider.Name), "", 1, "", true, 0, "")
	pdf.SetFont("Cantarell-Regular", "", 10)
	pdf.CellFormat(providerBlockWidth, lineHeight, fmt.Sprintf("%s: %s", translation.TaxID, config.Provider.ID), "", 1, "", true, 0, "")
	pdf.MultiCell(providerBlockWidth, lineHeight, fmt.Sprintf("%s: %s", translation.Address, config.Provider.Address), "", "", true)
	pdf.CellFormat(providerBlockWidth, lineHeight, fmt.Sprintf("%s: %s", translation.Bank, config.Provider.Bank), "", 1, "", true, 0, "")
	pdf.CellFormat(providerBlockWidth, lineHeight, fmt.Sprintf("IBAN: %s", config.Provider.IBAN), "", 1, "", true, 0, "")
	pdf.CellFormat(providerBlockWidth, lineHeight, fmt.Sprintf("SWIFT: %s", config.Provider.SWIFT), "", 1, "", true, 0, "")
	pdf.CellFormat(providerBlockWidth, lineHeight, fmt.Sprintf("%s %s", config.Provider.Phone, config.Provider.Email), "", 1, "", true, 0, "")

	// Invoice details block
	pdf.SetXY(providerBlockWidth+leftMargin, topMargin)
	pdf.SetTextColor(whiteColor.IntTuple())
	pdf.SetFillColor(primaryColor.IntTuple())
	pdf.SetFont("Cantarell-Bold", "", 16)
	pdf.CellFormat(invoiceBlockWidth, 10, strings.ToUpper(translation.Invoice)+" ", "", 2, "R", true, 0, "")
	pdf.SetFont("Cantarell-Regular", "", 10)
	now := time.Now()
	pdf.CellFormat(invoiceBlockWidth, lineHeight, fmt.Sprintf("%s: %s", translation.InvoiceID, now.Format("01-2006"))+" ", "", 2, "R", true, 0, "")
	pdf.CellFormat(invoiceBlockWidth, lineHeight*6, fmt.Sprintf("%s: %s", translation.InvoiceDate, now.Format("02.01.2006"))+" ", "", 1, "TR", true, 0, "")
	pdf.Ln(5)

	// BillTo information
	pdf.SetTextColor(primaryColor.IntTuple())
	pdf.SetFillColor(whiteColor.IntTuple())
	pdf.SetFont("Cantarell-Bold", "", 10)
	pdf.CellFormat(pageWidth, lineHeight, fmt.Sprintf("%s: %s", translation.BillTo, config.Company.Name), "", 1, "", true, 0, "")
	pdf.SetFont("Cantarell-Regular", "", 10)
	pdf.CellFormat(pageWidth, lineHeight, fmt.Sprintf("%s: %s", translation.Address, config.Company.Address), "", 1, "", true, 0, "")
	pdf.CellFormat(pageWidth, lineHeight, config.Company.Additional, "", 1, "", true, 0, "")
	pdf.Ln(5)

	// Invoice main table
	header := []string{translation.Quantity, translation.Description, translation.Price, "Total"}
	data := [][]string{
		{
			fmt.Sprintf("%.2f", config.Quantity),
			fmt.Sprintf("%s %s (%s)", translation.Agreement, config.Company.ContractID, config.Dates),
			fmt.Sprintf("%.2f €", config.Price),
			fmt.Sprintf("%.2f €", config.Quantity*config.Price),
		},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
	}
	drawTable(pdf, header, data)
	pdf.Ln(5)

	// Additional information
	a := strings.Split(translation.Additional, "\n")
	y_pos := pdf.GetY()
	for _, row := range a {
		pdf.CellFormat(additionalBlockWidth, lineHeight, row, "", 1, "C", false, 0, "")
	}

	// Total cost table
	// Titles
	pdf.SetXY(additionalBlockWidth+leftMargin, y_pos)
	pdf.SetCellMargin(2)
	pdf.CellFormat(30, lineHeight, translation.Subtotal, "", 2, "R", false, 0, "")
	pdf.CellFormat(30, lineHeight, translation.Tax, "", 2, "R", false, 0, "")
	pdf.CellFormat(30, lineHeight, translation.Shipping, "", 2, "R", false, 0, "")
	pdf.CellFormat(30, lineHeight, "Total", "", 0, "R", false, 0, "")
	// Numbers
	pdf.SetXY(additionalBlockWidth+leftMargin+30, y_pos)
	pdf.SetCellMargin(0)
	pdf.CellFormat(20, lineHeight, fmt.Sprintf("%.2f €", config.Quantity*config.Price), "B", 2, "R", false, 0, "")
	pdf.CellFormat(20, lineHeight, "0.0 €", "B", 2, "R", false, 0, "")
	pdf.CellFormat(20, lineHeight, "0.0 €", "B", 2, "R", false, 0, "")
	pdf.CellFormat(20, lineHeight, fmt.Sprintf("%.2f €", config.Quantity*config.Price), "B", 0, "R", false, 0, "")
}

func CreatePDF(config *Config) (string, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("Cantarell-Regular", "", fontCantarellRegular)
	pdf.AddUTF8FontFromBytes("Cantarell-Bold", "", fontCantarellBold)

	for _, lang := range config.Langs {
		translate, err := LoadTranslation(lang)
		if err != nil {
			return "", fmt.Errorf("unable to load translation for %s: %v", lang, err)
		}
		t, err := template.New("Additional").Parse(translate.Additional)
		if err != nil {
			return "", fmt.Errorf("unable to load template: %v", err)
		}
		var buf bytes.Buffer
		err = t.Execute(&buf, map[string]interface{}{"Contacts": fmt.Sprintf("%s <%s>", config.Provider.Name, config.Provider.Email)})
		if err != nil {
			return "", fmt.Errorf("unable to execute template: %v", err)
		}
		translate.Additional = buf.String()
		generatePage(pdf, config, translate)
	}
	now := time.Now()
	fileName := fmt.Sprintf("invoice-%s.pdf", now.Format("01-06"))
	return fileName, pdf.OutputFileAndClose(fileName)
}
