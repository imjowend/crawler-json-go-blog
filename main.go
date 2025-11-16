package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Post representa un post del blog
type Post struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// Lista de los primeros 25 posts (del más viejo al más nuevo)
var postSlugs = []string{
	"hello-world",
	"third-party-libraries-goprotobuf-and",
	"json-rpc-tale-of-interfaces",
	"new-talk-and-tutorials",
	"upcoming-google-io-go-events",
	"go-at-io-frequently-asked-questions",
	"go-programming-session-video-from",
	"gos-declaration-syntax",
	"share-memory-by-communicating",
	"defer-panic-and-recover",
	"go-wins-2010-bossie-award",
	"introducing-go-playground",
	"go-concurrency-patterns-timing-out-and",
	"real-go-projects-smarttwitter-and-webgo",
	"debugging-go-code-status-report",
	"go-one-year-ago-today",
	"go-slices-usage-and-internals",
	"json-and-go",
	"go-becomes-more-stable",
	"c-go-cgo",
	"gobs-of-data",
	"godoc-documenting-go-code",
	"introducing-gofix",
	"go-at-heroku",
	"go-and-google-app-engine",
}

func main() {
	posts := make([]Post, 0, 25)

	for i, slug := range postSlugs {
		fmt.Printf("Procesando post %d/25...\n", i+1)

		url := fmt.Sprintf("https://go.dev/blog/%s", slug)
		post, err := scrapePost(i+1, url, slug)
		if err != nil {
			fmt.Printf("Error procesando %s: %v\n", url, err)
			// Agregar post con información básica aunque falle
			post = Post{
				ID:      i + 1,
				Title:   "",
				Date:    "",
				URL:     url,
				Content: fmt.Sprintf("Error descargando contenido: %v", err),
			}
		}

		posts = append(posts, post)

		// Pequeña pausa para no sobrecargar el servidor
		time.Sleep(500 * time.Millisecond)
	}

	// Guardar en archivo JSON
	if err := saveToJSON(posts, "posts.json"); err != nil {
		fmt.Printf("Error guardando JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Todos los posts han sido guardados en posts.json")
}

func scrapePost(id int, url, slug string) (Post, error) {
	// Crear request con headers apropiados
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Post{}, fmt.Errorf("error creando request: %w", err)
	}

	// Agregar headers para evitar bloqueos
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	// Descargar la página
	resp, err := client.Do(req)
	if err != nil {
		return Post{}, fmt.Errorf("error descargando página: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Post{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Post{}, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	html := string(body)

	// Extraer título
	title := extractTitle(html)

	// Extraer fecha
	date := extractDate(html)

	// Extraer contenido
	content := extractContent(html)

	return Post{
		ID:      id,
		Title:   title,
		Date:    date,
		URL:     url,
		Content: content,
	}, nil
}

func extractTitle(html string) string {
	// Buscar el título en diferentes formatos comunes
	patterns := []string{
		`<h1[^>]*class="[^"]*article-title[^"]*"[^>]*>([^<]+)</h1>`,
		`<h1[^>]*>([^<]+)</h1>`,
		`<title>([^<]+)</title>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			// Limpiar sufijos comunes del título
			title = strings.TrimSuffix(title, " - The Go Programming Language")
			title = strings.TrimSuffix(title, " - The Go Blog")
			title = cleanHTML(title)
			if title != "" {
				return title
			}
		}
	}

	return "Sin título"
}

func extractDate(html string) string {
	// Buscar fecha en diferentes formatos
	patterns := []string{
		`<time[^>]*datetime="([^"]+)"`,
		`<span[^>]*class="[^"]*date[^"]*"[^>]*>([^<]+)</span>`,
		`<div[^>]*class="[^"]*date[^"]*"[^>]*>([^<]+)</div>`,
		`(\d{1,2}\s+\w+\s+\d{4})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			dateStr := strings.TrimSpace(matches[1])
			// Intentar parsear y convertir a formato YYYY-MM-DD
			normalized := normalizeDate(dateStr)
			if normalized != "" {
				return normalized
			}
		}
	}

	return ""
}

func normalizeDate(dateStr string) string {
	// Formatos comunes
	formats := []string{
		"2006-01-02",
		"02 January 2006",
		"January 2, 2006",
		"2 January 2006",
		"Jan 2, 2006",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}

	return dateStr
}

func extractContent(html string) string {
	// Extraer el contenido principal del artículo
	// Buscar el div/section del artículo
	patterns := []string{
		`<article[^>]*>([\s\S]*?)</article>`,
		`<div[^>]*class="[^"]*Article[^"]*"[^>]*>([\s\S]*?)</div>`,
		`<div[^>]*class="[^"]*article[^"]*"[^>]*>([\s\S]*?)</div>`,
		`<div[^>]*id="content"[^>]*>([\s\S]*?)</div>`,
		`<main[^>]*>([\s\S]*?)</main>`,
	}

	var content string
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			content = matches[1]
			break
		}
	}

	if content == "" {
		// Si no encontramos el contenedor específico, buscar después del h1
		re := regexp.MustCompile(`<h1[^>]*>.*?</h1>([\s\S]{0,50000})`)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			content = matches[1]
		}
	}

	// Si aún no hay contenido, tomar todo después del body
	if content == "" {
		re := regexp.MustCompile(`<body[^>]*>([\s\S]*?)</body>`)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			content = matches[1]
		}
	}

	// Limpiar HTML y extraer texto
	content = cleanHTML(content)
	content = strings.TrimSpace(content)

	// Limitar la longitud si es muy largo
	if len(content) > 15000 {
		content = content[:15000] + "..."
	}

	if content == "" {
		content = "No se pudo extraer el contenido"
	}

	return content
}

func cleanHTML(s string) string {
	// Remover scripts y styles
	re := regexp.MustCompile(`<script[^>]*>[\s\S]*?</script>`)
	s = re.ReplaceAllString(s, "")
	re = regexp.MustCompile(`<style[^>]*>[\s\S]*?</style>`)
	s = re.ReplaceAllString(s, "")

	// Remover comentarios HTML
	re = regexp.MustCompile(`<!--[\s\S]*?-->`)
	s = re.ReplaceAllString(s, "")

	// Convertir <br> y <p> en saltos de línea
	re = regexp.MustCompile(`<br[^>]*>`)
	s = re.ReplaceAllString(s, "\n")
	re = regexp.MustCompile(`</p>`)
	s = re.ReplaceAllString(s, "\n\n")

	// Remover todas las etiquetas HTML
	re = regexp.MustCompile(`<[^>]+>`)
	s = re.ReplaceAllString(s, " ")

	// Decodificar entidades HTML comunes
	replacements := map[string]string{
		"&nbsp;":   " ",
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&#39;":    "'",
		"&apos;":   "'",
		"&mdash;":  "—",
		"&ndash;":  "–",
		"&hellip;": "...",
	}

	for entity, replacement := range replacements {
		s = strings.ReplaceAll(s, entity, replacement)
	}

	// Limpiar espacios múltiples
	re = regexp.MustCompile(`[ \t]+`)
	s = re.ReplaceAllString(s, " ")

	// Limpiar múltiples saltos de línea
	re = regexp.MustCompile(`\n\s*\n\s*\n+`)
	s = re.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}

func saveToJSON(posts []Post, filename string) error {
	data, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo: %w", err)
	}

	return nil
}
