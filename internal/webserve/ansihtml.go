package webserve

import (
	"html"
	"strconv"
	"strings"
	"unicode/utf8"
)

type ansiStyleState struct {
	bold, italic, dim bool
	fg, bg            string
}

func (s *ansiStyleState) reset() {
	s.bold = false
	s.italic = false
	s.dim = false
	s.fg = ""
	s.bg = ""
}

func (s *ansiStyleState) hasStyle() bool {
	return s.bold || s.italic || s.dim || s.fg != "" || s.bg != ""
}

func (s *ansiStyleState) buildCSS() string {
	var style strings.Builder
	if s.bold {
		style.WriteString("font-weight:700;")
	}
	if s.italic {
		style.WriteString("font-style:italic;")
	}
	if s.dim {
		style.WriteString("opacity:0.65;")
	}
	if s.fg != "" {
		style.WriteString("color:")
		style.WriteString(s.fg)
		style.WriteByte(';')
	}
	if s.bg != "" {
		style.WriteString("background:")
		style.WriteString(s.bg)
		style.WriteByte(';')
	}
	return style.String()
}

// ANSIToHTML converts lipgloss/ANSI colored terminal text into HTML spans.
// Supports SGR reset, bold, italic, dim, 256-color, and truecolor fg/bg.
func ANSIToHTML(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s) + 64)

	var state ansiStyleState
	open := false

	flushOpen := func() {
		if open {
			b.WriteString("</span>")
			open = false
		}
	}
	openSpan := func() {
		flushOpen()
		css := state.buildCSS()
		if css == "" {
			return
		}
		b.WriteString(`<span style="`)
		b.WriteString(css)
		b.WriteString(`">`)
		open = true
	}
	ensureStyle := func() {
		if !open && state.hasStyle() {
			openSpan()
		}
	}

	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			end := i + 2
			for end < len(s) && !((s[end] >= '@' && s[end] <= 'Z') || (s[end] >= 'a' && s[end] <= 'z')) {
				end++
			}
			if end >= len(s) {
				break
			}
			cmd := s[end]
			params := s[i+2 : end]
			i = end + 1
			if cmd == 'm' {
				flushOpen()
				parseSGRParams(params, &state)
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		ensureStyle()
		b.WriteString(html.EscapeString(string(r)))
		i += size
	}
	flushOpen()
	return b.String()
}

func parseSGRParams(params string, state *ansiStyleState) {
	if params == "" || params == "0" {
		state.reset()
		return
	}
	parts := strings.Split(params, ";")
	for j := 0; j < len(parts); j++ {
		n, err := strconv.Atoi(parts[j])
		if err != nil {
			continue
		}
		switch {
		case n == 0:
			state.reset()
		case n == 1:
			state.bold = true
		case n == 2:
			state.dim = true
		case n == 3:
			state.italic = true
		case n == 22:
			state.bold, state.dim = false, false
		case n == 23:
			state.italic = false
		case n == 39:
			state.fg = ""
		case n == 49:
			state.bg = ""
		case n >= 30 && n <= 37:
			state.fg = basicANSIColor(n - 30)
		case n >= 90 && n <= 97:
			state.fg = brightANSIColor(n - 90)
		case n >= 40 && n <= 47:
			state.bg = basicANSIColor(n - 40)
		case n >= 100 && n <= 107:
			state.bg = brightANSIColor(n - 100)
		case n == 38 || n == 48:
			parseExtendedColor(n == 38, parts, &j, state)
		}
	}
}

func parseExtendedColor(isFG bool, parts []string, j *int, state *ansiStyleState) {
	if *j+1 >= len(parts) {
		return
	}
	mode, _ := strconv.Atoi(parts[*j+1])
	*j++
	switch mode {
	case 5:
		if *j+1 >= len(parts) {
			return
		}
		idx, _ := strconv.Atoi(parts[*j+1])
		*j++
		c := ansi256(idx)
		if isFG {
			state.fg = c
		} else {
			state.bg = c
		}
	case 2:
		if *j+3 >= len(parts) {
			return
		}
		r, _ := strconv.Atoi(parts[*j+1])
		g, _ := strconv.Atoi(parts[*j+2])
		bl, _ := strconv.Atoi(parts[*j+3])
		*j += 3
		c := rgb(r, g, bl)
		if isFG {
			state.fg = c
		} else {
			state.bg = c
		}
	}
}

func rgb(r, g, b int) string {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return v
	}
	return "#" + hex2(clamp(r)) + hex2(clamp(g)) + hex2(clamp(b))
}

func hex2(v int) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[v>>4], digits[v&0xf]})
}

func basicANSIColor(n int) string {
	switch n {
	case 0:
		return "#000000"
	case 1:
		return "#cd0000"
	case 2:
		return "#00cd00"
	case 3:
		return "#cdcd00"
	case 4:
		return "#0000ee"
	case 5:
		return "#cd00cd"
	case 6:
		return "#00cdcd"
	default:
		return "#e5e5e5"
	}
}

func brightANSIColor(n int) string {
	switch n {
	case 0:
		return "#7f7f7f"
	case 1:
		return "#ff0000"
	case 2:
		return "#00ff00"
	case 3:
		return "#ffff00"
	case 4:
		return "#5c5cff"
	case 5:
		return "#ff00ff"
	case 6:
		return "#00ffff"
	default:
		return "#ffffff"
	}
}

func ansi256(n int) string {
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	if n < 16 {
		palette := []string{
			"#000000", "#800000", "#008000", "#808000",
			"#000080", "#800080", "#008080", "#c0c0c0",
			"#808080", "#ff0000", "#00ff00", "#ffff00",
			"#0000ff", "#ff00ff", "#00ffff", "#ffffff",
		}
		return palette[n]
	}
	if n >= 232 {
		v := 8 + (n-232)*10
		return rgb(v, v, v)
	}
	n -= 16
	r := n / 36
	g := (n % 36) / 6
	b := n % 6
	levels := []int{0, 95, 135, 175, 215, 255}
	return rgb(levels[r], levels[g], levels[b])
}
