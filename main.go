package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"regexp"

	"github.com/gorilla/mux"
)

func main() {
	initializeBlogPosts()
	router := mux.NewRouter()
	initializeRoutes(router)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func formatFullDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}

var funcMap = template.FuncMap{
	"avail":       avail,
	"add":         func(a, b int) int { return a + b },
	"addFloat":    func(a, b float64) float64 { return a + b },
	"currentYear": func() int { return time.Now().Year() },
	"formatDate": func(dateStr string) string {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr
		}
		return t.Format("January 2, 2006")
	},
	"formatFullDate": formatFullDate,
	"formatActivityType": func(activityType string) string {
		if activityType == "" {
			return ""
		}
		// Convert camelCase or PascalCase to spaced words
		re := regexp.MustCompile(`([a-z])([A-Z])`)
		return re.ReplaceAllString(activityType, "$1 $2")
	},
	"htmlSafe":       func(html string) template.HTML { return template.HTML(html) },
	"optimizeImages": func(html string) template.HTML { return template.HTML(optimizeImagesInContent(html)) },
	"getTagCount": func(tag string) int {
		count, ok := tagCountMap[tag]
		if !ok {
			return 0
		}
		return count
	},
	"slugifyTitle": slugifyTitle,
	"truncateTitle": func(title string) string {
		words := strings.Fields(title)
		if len(words) <= 4 {
			return title
		}
		truncated := strings.Join(words[:4], " ")
		truncated = strings.TrimRight(truncated, ".,!?:;")
		return truncated + "..."
	},
	"tagUrl":  func(tag string) string { return "/all?tag=" + url.QueryEscape(tag) },
	"yearUrl": func(year int) string { return "/all?year=" + strconv.Itoa(year) },
	"getYearCount": func(year int) int {
		posts, ok := blogPostYearMap[year]
		if !ok {
			return 0
		}
		return len(posts)
	},
	"contains": func(s, substr string) bool { return strings.Contains(s, substr) },
	"split":    func(s, sep string) []string { return strings.Split(s, sep) },
	"atoi":     func(s string) int { i, _ := strconv.Atoi(strings.TrimSpace(s)); return i },
	"atof":     func(s string) float64 { f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64); return f },
	"mul":      func(a, b int) int { return a * b },
	"mulFloat": func(a, b float64) float64 { return a * b },
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"mod": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a % b
	},
	"replace": func(s, old, new string) string { return strings.ReplaceAll(s, old, new) },
	"printf":  func(format string, args ...interface{}) string { return fmt.Sprintf(format, args...) },
	"len": func(slice interface{}) int {
		v := reflect.ValueOf(slice)
		if v.Kind() == reflect.Slice {
			return v.Len()
		}
		return 0
	},
	"commaf": func(val float64, decimals int) string {
		format := "%0." + strconv.Itoa(decimals) + "f"
		num := fmt.Sprintf(format, val)
		parts := strings.Split(num, ".")
		intPart := parts[0]
		var out []byte
		for i, c := range intPart {
			if i != 0 && (len(intPart)-i)%3 == 0 {
				out = append(out, ',')
			}
			out = append(out, byte(c))
		}
		if len(parts) > 1 {
			return string(out) + "." + parts[1]
		}
		return string(out)
	},
	"commai": func(val int) string {
		in := strconv.Itoa(val)
		n := len(in)
		if n <= 3 {
			return in
		}
		rem := n % 3
		if rem == 0 {
			rem = 3
		}
		out := in[:rem]
		for i := rem; i < n; i += 3 {
			out += "," + in[i:i+3]
		}
		return out
	},
	"join": func(elems []string, sep string) string { return strings.Join(elems, sep) },
	"title": func(s string) string {
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	},
	"detectCodeLanguage": func(code string) string {
		lines := strings.Split(code, "\n")
		if len(lines) == 0 {
			return "text"
		}
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "#!/") {
			if strings.Contains(firstLine, "python") {
				return "python"
			}
			if strings.Contains(firstLine, "bash") || strings.Contains(firstLine, "sh") {
				return "bash"
			}
			if strings.Contains(firstLine, "node") {
				return "javascript"
			}
		}
		if strings.Contains(firstLine, "package main") {
			return "go"
		}
		if strings.Contains(firstLine, "import ") && strings.Contains(firstLine, "(") {
			return "go"
		}
		if strings.Contains(firstLine, "function ") || strings.Contains(firstLine, "const ") || strings.Contains(firstLine, "let ") {
			return "javascript"
		}
		if strings.Contains(firstLine, "def ") || strings.Contains(firstLine, "import ") {
			return "python"
		}
		if strings.Contains(firstLine, "<!DOCTYPE") || strings.Contains(firstLine, "<html") {
			return "html"
		}
		if strings.Contains(firstLine, "{") && strings.Contains(firstLine, "}") {
			return "json"
		}
		return "text"
	},
}
