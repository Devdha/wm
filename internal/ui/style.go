package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Color scheme for the TUI
var (
	// Primary colors
	Success = color.New(color.FgGreen, color.Bold)
	Error   = color.New(color.FgRed, color.Bold)
	Warning = color.New(color.FgYellow, color.Bold)
	Info    = color.New(color.FgCyan)
	Muted   = color.New(color.Faint)
	Bold    = color.New(color.Bold)

	// Accent colors
	Primary = color.New(color.FgBlue, color.Bold)
	Accent  = color.New(color.FgMagenta)
)

// Icons using Unicode characters
const (
	IconCheck    = "✓"
	IconCross    = "✗"
	IconArrow    = "→"
	IconBolt     = "⚡"
	IconPackage  = "📦"
	IconRocket   = "🚀"
	IconFolder   = "📁"
	IconBranch   = "🌿"
	IconSparkle  = "✨"
	IconWarning  = "⚠"
	IconInfo     = "ℹ"
	IconTree     = "├─"
	IconTreeEnd  = "└─"
)

// Box drawing characters
const (
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxCross       = "┼"
	BoxTeeDown     = "┬"
	BoxTeeUp       = "┴"
	BoxTeeRight    = "├"
	BoxTeeLeft     = "┤"
)

// Styled prints
func PrintSuccess(message string) {
	Success.Print(IconCheck + " ")
	fmt.Println(message)
}

func PrintError(message string) {
	Error.Print(IconCross + " ")
	fmt.Println(message)
}

func PrintWarning(message string) {
	Warning.Print(IconWarning + " ")
	fmt.Println(message)
}

func PrintInfo(message string) {
	Info.Print(IconInfo + " ")
	fmt.Println(message)
}

func PrintStep(icon, message string) {
	fmt.Print(icon + " ")
	Bold.Println(message)
}

func PrintSubStep(message string) {
	fmt.Print("  " + IconTree + " ")
	fmt.Println(message)
}

func PrintSubStepEnd(message string) {
	fmt.Print("  " + IconTreeEnd + " ")
	fmt.Println(message)
}

// DrawBox creates a simple box around text
func DrawBox(title string, width int) {
	if width < len(title)+4 {
		width = len(title) + 4
	}

	padding := width - len(title) - 2
	leftPad := padding / 2
	rightPad := padding - leftPad

	// Top border
	fmt.Print(BoxTopLeft)
	fmt.Print(strings.Repeat(BoxHorizontal, width))
	fmt.Println(BoxTopRight)

	// Title line
	fmt.Print(BoxVertical)
	fmt.Print(strings.Repeat(" ", leftPad))
	Bold.Print(title)
	fmt.Print(strings.Repeat(" ", rightPad))
	fmt.Println(BoxVertical)

	// Bottom border
	fmt.Print(BoxBottomLeft)
	fmt.Print(strings.Repeat(BoxHorizontal, width))
	fmt.Println(BoxBottomRight)
}

// Table creates a formatted table with box drawing characters
type Table struct {
	Headers []string
	Rows    [][]string
	widths  []int
}

func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
		Rows:    make([][]string, 0),
		widths:  make([]int, len(headers)),
	}
}

func (t *Table) AddRow(cells ...string) {
	if len(cells) != len(t.Headers) {
		return
	}
	t.Rows = append(t.Rows, cells)
}

func (t *Table) Render() {
	// Calculate column widths
	for i, header := range t.Headers {
		t.widths[i] = len(header)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if len(cell) > t.widths[i] {
				t.widths[i] = len(cell)
			}
		}
	}

	// Top border
	t.printBorder(BoxTopLeft, BoxTeeDown, BoxTopRight)

	// Headers
	fmt.Print(BoxVertical)
	for i, header := range t.Headers {
		fmt.Print(" ")
		Bold.Print(header)
		fmt.Print(strings.Repeat(" ", t.widths[i]-len(header)+1))
		fmt.Print(BoxVertical)
	}
	fmt.Println()

	// Middle border
	t.printBorder(BoxTeeRight, BoxCross, BoxTeeLeft)

	// Rows
	for _, row := range t.Rows {
		fmt.Print(BoxVertical)
		for i, cell := range row {
			fmt.Print(" ")
			fmt.Print(cell)
			fmt.Print(strings.Repeat(" ", t.widths[i]-len(cell)+1))
			fmt.Print(BoxVertical)
		}
		fmt.Println()
	}

	// Bottom border
	t.printBorder(BoxBottomLeft, BoxTeeUp, BoxBottomRight)
}

func (t *Table) printBorder(left, middle, right string) {
	fmt.Print(left)
	for i, width := range t.widths {
		fmt.Print(strings.Repeat(BoxHorizontal, width+2))
		if i < len(t.widths)-1 {
			fmt.Print(middle)
		}
	}
	fmt.Println(right)
}

// Progress indicator helpers
func PrintProgress(current, total int, message string) {
	Info.Printf("[%d/%d] ", current, total)
	fmt.Println(message)
}

// Muted text helper
func Mute(text string) string {
	return Muted.Sprint(text)
}
