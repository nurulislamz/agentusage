package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// getRepoRoot locates the repository root directory by walking up from the current working directory.
func getRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repository root containing go.mod from %s", dir)
		}
		dir = parent
	}
}

func readCommandFlowDiagramsDoc(t *testing.T) (string, string) {
	t.Helper()
	root := getRepoRoot(t)
	docPath := filepath.Join(root, "docs", "COMMAND_FLOW_DIAGRAMS.md")
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", docPath, err)
	}
	return root, string(data)
}
func extractPlantUMLBlocks(content string) []string {
	re := regexp.MustCompile("(?s)```plantuml\\s*\\n(@startuml.*?@enduml)\\s*\\n```")
	matches := re.FindAllStringSubmatch(content, -1)
	var blocks []string
	for _, m := range matches {
		if len(m) > 1 {
			blocks = append(blocks, m[1])
		}
	}
	return blocks
}

func extractSVGImageRefs(content string) []string {
	re := regexp.MustCompile(`!\[([^\]]*)\]\((diagrams/[^)]+\.svg)\)`)
	matches := re.FindAllStringSubmatch(content, -1)
	var refs []string
	for _, m := range matches {
		if len(m) > 2 {
			refs = append(refs, m[2])
		}
	}
	return refs
}

func TestCommandFlowDiagrams_NoMermaidBlocks(t *testing.T) {
	_, content := readCommandFlowDiagramsDoc(t)

	if strings.Contains(content, "```mermaid") {
		t.Errorf("docs/COMMAND_FLOW_DIAGRAMS.md still contains ```mermaid blocks; all diagrams must be PlantUML")
	}
}

func TestCommandFlowDiagrams_PlantUMLBlocksPresent(t *testing.T) {
	_, content := readCommandFlowDiagramsDoc(t)

	blocks := extractPlantUMLBlocks(content)
	const expectedCount = 14
	if len(blocks) != expectedCount {
		t.Fatalf("expected exactly %d PlantUML diagram blocks, found %d", expectedCount, len(blocks))
	}
}

func TestCommandFlowDiagrams_PlantUMLSyntaxValidity(t *testing.T) {
	_, content := readCommandFlowDiagramsDoc(t)
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
			var blockDepth int // for alt, opt, loop, par, group

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

func TestCommandFlowDiagrams_RenderedSVGsIntegrity(t *testing.T) {
	root, content := readCommandFlowDiagramsDoc(t)
	svgRefs := extractSVGImageRefs(content)

	if len(svgRefs) != 14 {
		t.Fatalf("expected 14 diagram SVG image references, found %d", len(svgRefs))
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

func TestCommandFlowDiagrams_PlantUMLCompiler(t *testing.T) {
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

	_, content := readCommandFlowDiagramsDoc(t)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
