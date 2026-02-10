package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coolamit/mermaid-cli/internal/config"
	"github.com/coolamit/mermaid-cli/internal/icons"
	"github.com/coolamit/mermaid-cli/internal/markdown"
	"github.com/coolamit/mermaid-cli/internal/renderer"
	"github.com/spf13/cobra"
)

// Version is set at build time.
var Version = "dev"

// Flags holds all CLI flag values.
type Flags struct {
	Input                 string
	Output                string
	Artefacts             string
	Theme                 string
	Width                 int
	Height                int
	BackgroundColor       string
	OutputFormat          string
	Scale                 int
	PdfFit                bool
	SvgFit                bool
	SVGId                 string
	ConfigFile            string
	CSSFile               string
	PuppeteerConfigFile   string
	IconPacks             []string
	IconPacksNamesAndUrls []string
	Quiet                 bool
}

// NewRootCommand creates the cobra root command with all flags.
func NewRootCommand() *cobra.Command {
	flags := &Flags{}

	cmd := &cobra.Command{
		Use:     "mmd-cli",
		Short:   "Mermaid CLI - Generate diagrams from mermaid definitions",
		Long:    "A CLI tool to convert mermaid diagram definitions into SVG, PNG, and PDF files.",
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(flags)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Define flags to match the official mermaid-cli exactly
	cmd.Flags().StringVarP(&flags.Input, "input", "i", "", "Input mermaid file. Files ending in .md will be treated as Markdown. Use `-` to read from stdin.")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "Output file. It should be either md, svg, png, pdf or use `-` for stdout. Default: input + \".svg\"")
	cmd.Flags().StringVarP(&flags.Artefacts, "artefacts", "a", "", "Output artefacts path. Only used with Markdown input.")
	cmd.Flags().StringVarP(&flags.Theme, "theme", "t", "default", "Theme of the chart (default, forest, dark, neutral)")
	cmd.Flags().IntVarP(&flags.Width, "width", "w", 800, "Width of the page")
	cmd.Flags().IntVarP(&flags.Height, "height", "H", 600, "Height of the page")
	cmd.Flags().StringVarP(&flags.BackgroundColor, "backgroundColor", "b", "white", "Background color for pngs/svgs (not pdfs). Example: transparent, red, '#F0F0F0'.")
	cmd.Flags().StringVarP(&flags.OutputFormat, "outputFormat", "e", "", "Output format for the generated image (svg, png, pdf). Default: from output file extension")
	cmd.Flags().IntVarP(&flags.Scale, "scale", "s", 1, "Scale factor")
	cmd.Flags().BoolVarP(&flags.PdfFit, "pdfFit", "f", false, "Scale PDF to fit chart")
	cmd.Flags().BoolVar(&flags.SvgFit, "svgFit", false, "Set SVG dimensions to match diagram size (for standalone viewing)")
	cmd.Flags().StringVarP(&flags.SVGId, "svgId", "I", "", "The id attribute for the SVG element to be rendered")
	cmd.Flags().StringVarP(&flags.ConfigFile, "configFile", "c", "", "JSON configuration file for mermaid")
	cmd.Flags().StringVarP(&flags.CSSFile, "cssFile", "C", "", "CSS file for the page")
	cmd.Flags().StringVarP(&flags.PuppeteerConfigFile, "puppeteerConfigFile", "p", "", "JSON configuration file for the browser")
	cmd.Flags().StringSliceVar(&flags.IconPacks, "iconPacks", nil, "Icon packs to use, e.g. @iconify-json/logos")
	cmd.Flags().StringSliceVar(&flags.IconPacksNamesAndUrls, "iconPacksNamesAndUrls", nil, "Icon packs with name#url format")
	cmd.Flags().BoolVarP(&flags.Quiet, "quiet", "q", false, "Suppress log output")

	return cmd
}

// info logs a message unless quiet mode is enabled.
func info(quiet bool, format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// errorExit prints an error message in red and exits.
func errorExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m\n%s\n\033[0m", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func run(flags *Flags) error {
	input := flags.Input
	output := flags.Output
	outputFormat := flags.OutputFormat
	quiet := flags.Quiet

	// Validate input
	if input == "" {
		info(false, "No input file specified, reading from stdin. "+
			"If you want to specify an input file, please use `-i <input>.` "+
			"You can use `-i -` to read from stdin and to suppress this warning.")
	} else if input == "-" {
		// stdin mode, suppress warning
		input = ""
	} else if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file %q doesn't exist", input)
	}

	// Determine output
	if output == "" {
		if outputFormat != "" {
			if input != "" {
				output = input + "." + outputFormat
			} else {
				output = "out." + outputFormat
			}
		} else {
			if input != "" {
				output = input + ".svg"
			} else {
				output = "out.svg"
			}
		}
	} else if output == "-" {
		output = "/dev/stdout"
		quiet = true
		if outputFormat == "" {
			outputFormat = "svg"
			info(false, "No output format specified, using svg. "+
				"If you want to specify an output format and suppress this warning, "+
				"please use `-e <format>.`")
		}
	} else {
		validExt := regexp.MustCompile(`\.(?:svg|png|pdf|md|markdown)$`)
		if !validExt.MatchString(output) {
			return fmt.Errorf("output file must end with \".md\"/\".markdown\", \".svg\", \".png\" or \".pdf\"")
		}
	}

	// Validate artefacts
	if flags.Artefacts != "" {
		if input == "" || !regexp.MustCompile(`\.(?:md|markdown)$`).MatchString(input) {
			return fmt.Errorf("artefacts [-a|--artefacts] path can only be used with Markdown input file")
		}
		if err := os.MkdirAll(flags.Artefacts, 0755); err != nil {
			return fmt.Errorf("failed to create artefacts directory: %w", err)
		}
	}

	// Check output directory exists
	if output != "/dev/stdout" {
		outputDir := filepath.Dir(output)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			return fmt.Errorf("output directory %q/ doesn't exist", outputDir)
		}
	}

	// Determine output format from extension
	if outputFormat == "" {
		ext := strings.TrimPrefix(filepath.Ext(output), ".")
		if ext == "md" || ext == "markdown" {
			outputFormat = "svg"
		} else {
			outputFormat = ext
		}
	}

	validFormats := regexp.MustCompile(`^(?:svg|png|pdf)$`)
	if !validFormats.MatchString(outputFormat) {
		return fmt.Errorf("output format must be one of \"svg\", \"png\" or \"pdf\"")
	}

	// Load configs
	mermaidConfig, err := config.LoadMermaidConfig(flags.ConfigFile, flags.Theme)
	if err != nil {
		return err
	}

	browserConfig, err := config.LoadBrowserConfig(flags.PuppeteerConfigFile)
	if err != nil {
		return err
	}

	css, err := config.LoadCSSFile(flags.CSSFile)
	if err != nil {
		return err
	}

	// Collect icon packs
	var allIconPacks []icons.IconPack
	if len(flags.IconPacks) > 0 {
		allIconPacks = append(allIconPacks, icons.ParseIconPacks(flags.IconPacks)...)
	}
	if len(flags.IconPacksNamesAndUrls) > 0 {
		allIconPacks = append(allIconPacks, icons.ParseIconPacksNamesAndUrls(flags.IconPacksNamesAndUrls)...)
	}

	// Build render options
	renderOpts := renderer.RenderOpts{
		MermaidConfig:   mermaidConfig,
		BackgroundColor: flags.BackgroundColor,
		CSS:             css,
		SVGId:           flags.SVGId,
		Width:           flags.Width,
		Height:          flags.Height,
		Scale:           flags.Scale,
		PdfFit:          flags.PdfFit,
		SvgFit:          flags.SvgFit,
		IconPacks:       allIconPacks,
	}

	// Read input
	var definition string
	if input != "" {
		data, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		definition = string(data)
	} else {
		data, err := readStdin()
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		definition = string(data)
	}

	// Set up renderer
	browser := renderer.NewBrowser(browserConfig)
	r := renderer.NewRenderer(browser)
	defer r.Close()

	ctx := context.Background()

	// Handle markdown input
	if input != "" && regexp.MustCompile(`\.(?:md|markdown)$`).MatchString(input) {
		if output == "/dev/stdout" {
			return fmt.Errorf("cannot use `stdout` with markdown input")
		}

		diagrams := markdown.ExtractDiagrams(definition)

		if len(diagrams) > 0 {
			info(quiet, "Found %d mermaid charts in Markdown input", len(diagrams))
		} else {
			info(quiet, "No mermaid charts found in Markdown input")
		}

		imageRefs := make([]markdown.ImageRef, 0, len(diagrams))

		for _, diagram := range diagrams {
			// Build numbered output filename
			ext := filepath.Ext(output)
			base := strings.TrimSuffix(output, ext)
			// If output is .md/.markdown, use outputFormat extension for images
			imgExt := ext
			if ext == ".md" || ext == ".markdown" {
				imgExt = "." + outputFormat
			}
			outputFile := fmt.Sprintf("%s-%d%s", base, diagram.Index, imgExt)

			if flags.Artefacts != "" {
				outputFile = filepath.Join(flags.Artefacts, filepath.Base(outputFile))
			}

			// Calculate relative path from output dir
			outputDir := filepath.Dir(filepath.Clean(output))
			relPath, err := filepath.Rel(outputDir, filepath.Clean(outputFile))
			if err != nil {
				relPath = outputFile
			}
			outputFileRelative := "./" + relPath

			result, err := r.Render(ctx, diagram.Definition, outputFormat, renderOpts)
			if err != nil {
				return fmt.Errorf("failed to render diagram %d: %w", diagram.Index, err)
			}

			if err := os.WriteFile(outputFile, result.Data, 0644); err != nil {
				return fmt.Errorf("failed to write output file %q: %w", outputFile, err)
			}

			info(quiet, " ✅ %s", outputFileRelative)

			imageRefs = append(imageRefs, markdown.ImageRef{
				URL:   outputFileRelative,
				Alt:   result.Desc,
				Title: result.Title,
			})
		}

		// If output is markdown, replace code blocks with image references
		if regexp.MustCompile(`\.(?:md|markdown)$`).MatchString(output) {
			outContent := markdown.ReplaceDiagrams(definition, imageRefs)
			if err := os.WriteFile(output, []byte(outContent), 0644); err != nil {
				return fmt.Errorf("failed to write markdown output: %w", err)
			}
			info(quiet, " ✅ %s", output)
		}
	} else {
		// Single diagram rendering
		info(quiet, "Generating single mermaid chart")

		result, err := r.Render(ctx, definition, outputFormat, renderOpts)
		if err != nil {
			return err
		}

		if output == "/dev/stdout" {
			if _, err := os.Stdout.Write(result.Data); err != nil {
				return fmt.Errorf("failed to write to stdout: %w", err)
			}
		} else {
			if err := os.WriteFile(output, result.Data, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			info(quiet, " ✅ %s", output)
		}
	}

	return nil
}

// readStdin reads all data from stdin.
func readStdin() ([]byte, error) {
	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
	}
	return data, nil
}
