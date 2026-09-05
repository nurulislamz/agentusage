package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func readUserFlowDiagramDoc(t *testing.T) (string, string) {
	t.Helper()
	root := getRepoRoot(t)
	docPath := filepath.Join(root, "docs", "user-flow-diagram.md")
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", docPath, err)
	}
	return root, string(data)
}

func TestUserFlowDiagram_NoMermaidBlocks(t *testing.T) {
	_, content := readUserFlowDiagramDoc(t)

	if strings.Contains(content, "```mermaid") {
		t.Errorf("docs/user-flow-diagram.md contains ```mermaid blocks; all diagrams must be PlantUML")
	}
}

func TestUserFlowDiagram_PlantUMLBlocksPresent(t *testing.T) {
	_, content := readUserFlowDiagramDoc(t)

	blocks := extractPlantUMLBlocks(content)
	const expectedCount = 16
	if len(blocks) != expectedCount {
		t.Fatalf("expected exactly %d PlantUML diagram blocks, found %d", expectedCount, len(blocks))
	}
}

func TestUserFlowDiagram_PlantUMLSyntaxValidity(t *testing.T) {
	_, content := readUserFlowDiagramDoc(t)
	blocks := extractPlantUMLBlocks(content)

	for i, block := range blocks {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			lines := strings.Split(block, "\n")
			if len(lines) == 0 {
				t.Fatalf("diagram %d is empty", i)
			}

			if !strings.HasPrefix(strings.TrimSpace(lines[0]), "@startuml") {
				t.Errorf("diagram %d does not start with @startuml", i)
			}
			if !strings.HasSuffix(strings.TrimSpace(lines[len(lines)-1]), "@enduml") {
				t.Errorf("diagram %d does not end with @enduml", i)
			}

			var boxCount int
			var blockDepth int

			for lineNum, rawLine := range lines {
				line := strings.TrimSpace(rawLine)
				if strings.HasPrefix(line, "box ") {
					boxCount++
				}
				if line == "end box" {
					boxCount--
					if boxCount < 0 {
						t.Errorf("diagram %d line %d: unmatched 'end box'", i, lineNum+1)
					}
				}

				if strings.HasPrefix(line, "alt ") ||
					strings.HasPrefix(line, "opt ") ||
					strings.HasPrefix(line, "loop ") ||
					strings.HasPrefix(line, "par ") ||
					strings.HasPrefix(line, "group ") {
					blockDepth++
				}
				if line == "end" {
					blockDepth--
					if blockDepth < 0 {
						t.Errorf("diagram %d line %d: unmatched 'end'", i, lineNum+1)
					}
				}
			}

			if boxCount != 0 {
				t.Errorf("diagram %d has %d unclosed 'box' declarations", i, boxCount)
			}
			if blockDepth != 0 {
				t.Errorf("diagram %d has %d unclosed blocks (alt/opt/loop/par/group)", i, blockDepth)
			}
		})
	}
}

func TestUserFlowDiagram_RenderedSVGsIntegrity(t *testing.T) {
	root, content := readUserFlowDiagramDoc(t)
	svgRefs := extractSVGImageRefs(content)

	const expectedCount = 16
	if len(svgRefs) != expectedCount {
		t.Fatalf("expected %d diagram SVG image references, found %d", expectedCount, len(svgRefs))
	}

	for _, relPath := range svgRefs {
		t.Run(relPath, func(t *testing.T) {
			fullPath := filepath.Join(root, "docs", relPath)
			info, err := os.Stat(fullPath)
			if err != nil {
				t.Fatalf("diagram asset %s does not exist: %v", fullPath, err)
			}
			if info.Size() < 500 {
				t.Errorf("diagram asset %s is unexpectedly small (%d bytes)", fullPath, info.Size())
			}

			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Fatalf("failed reading %s: %v", fullPath, err)
			}
			svgContent := string(data)

			if !strings.Contains(svgContent, "<svg") || !strings.Contains(svgContent, "</svg>") {
				t.Errorf("file %s is not valid SVG", fullPath)
			}

			if strings.Contains(svgContent, "Syntax Error") {
				t.Errorf("file %s contains PlantUML syntax error rendering", fullPath)
			}
		})
	}
}

func TestUserFlowDiagram_PlantUMLCompiler(t *testing.T) {
	jarPath := os.Getenv("PLANTUML_JAR")
	if jarPath == "" {
		jarPath = "/tmp/plantuml.jar"
	}
	if _, err := os.Stat(jarPath); err != nil {
		t.Skipf("PlantUML jar not found at %s; skipping compiler test", jarPath)
	}

	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java executable not found; skipping PlantUML compiler test")
	}

	_, content := readUserFlowDiagramDoc(t)
	blocks := extractPlantUMLBlocks(content)

	for i, block := range blocks {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			cmd := exec.Command("java", "-jar", jarPath, "-tsvg", "-pipe")
			cmd.Stdin = strings.NewReader(block)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("PlantUML compiler failed on diagram %d: %v\nStderr: %s", i, err, stderr.String())
			}

			output := stdout.String()
			if !strings.Contains(output, "<svg") || strings.Contains(output, "Syntax Error") {
				t.Errorf("PlantUML output for diagram %d is invalid SVG: %s", i, output[:min(len(output), 300)])
			}
		})
	}
}
