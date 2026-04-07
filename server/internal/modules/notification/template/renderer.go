package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type Renderer struct {
	templates map[string]map[string]*template.Template // [eventType][lang] -> template
}

func NewRenderer(emailsDir string) (*Renderer, error) {
	templates := make(map[string]map[string]*template.Template)

	entries, err := os.ReadDir(emailsDir)
	if err != nil {
		return nil, fmt.Errorf("read emails dir %s: %w", emailsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		eventType := entry.Name()
		templates[eventType] = make(map[string]*template.Template)

		langFiles, err := filepath.Glob(filepath.Join(emailsDir, eventType, "*.html"))
		if err != nil {
			return nil, fmt.Errorf("glob %s templates: %w", eventType, err)
		}

		for _, f := range langFiles {
			lang := strings.TrimSuffix(filepath.Base(f), ".html")
			tmpl, err := template.ParseFiles(f)
			if err != nil {
				return nil, fmt.Errorf("parse template %s: %w", f, err)
			}
			templates[eventType][lang] = tmpl
		}
	}

	return &Renderer{templates: templates}, nil
}

func (r *Renderer) Render(eventType, lang string, data any) (string, string, error) {
	langMap, ok := r.templates[eventType]
	if !ok {
		return "", "", fmt.Errorf("unknown email event type: %s", eventType)
	}

	tmpl, ok := langMap[lang]
	if !ok {
		tmpl, ok = langMap["en"]
		if !ok {
			return "", "", fmt.Errorf("no template for event %s lang %s (no en fallback)", eventType, lang)
		}
	}

	var subjectBuf, bodyBuf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&subjectBuf, "subject", data); err != nil {
		return "", "", fmt.Errorf("render subject for %s/%s: %w", eventType, lang, err)
	}

	if err := tmpl.ExecuteTemplate(&bodyBuf, "body", data); err != nil {
		return "", "", fmt.Errorf("render body for %s/%s: %w", eventType, lang, err)
	}

	return strings.TrimSpace(subjectBuf.String()), strings.TrimSpace(bodyBuf.String()), nil
}
