package utils

import (
	"html/template"
	"net/http"
	"path"

	"github.com/gorilla/csrf"
)

var templateMap map[string]*template.Template = map[string]*template.Template{}

//ViewData is the Type for the Views
type ViewData map[string]interface{}

//ContextKeyType for Request Context
type ContextKeyType struct{}

//GetViewData for the RenderView
func GetViewData(r *http.Request) ViewData {
	data := ViewData{}
	usr := r.Context().Value(ContextKeyType{})
	if usr != nil {
		data["CurrentUser"] = usr
	}
	data[csrf.TemplateTag] = csrf.TemplateField(r)
	return data
}

//GetPageTemplate to Writer
func GetPageTemplate(basePath string, viewsPath string) *template.Template {
	//templates in executive order
	viewsPaths := []string{
		viewsPath,
		"layouts/shell.layout.html",
		"layouts/page/page.layout.html",
		"layouts/page/head-nav.layout.html",
		"layouts/page/side-nav.layout.html",
	}

	views := []string{}
	for _, p := range viewsPaths {
		views = append(views, path.Join(basePath, p))
	}
	return getCachedTemplate(views, viewsPath)
}

//GetPlainTemplate to Writer
func GetPlainTemplate(basePath string, viewsPath string) *template.Template {
	//templates in executive order
	viewsPaths := []string{
		viewsPath,
		"layouts/shell.layout.html",
		"layouts/plain/plain.layout.html",
	}

	views := []string{}
	for _, p := range viewsPaths {
		views = append(views, path.Join(basePath, p))
	}
	return getCachedTemplate(views, viewsPath)
}

func getCachedTemplate(views []string, viewsPath string) *template.Template {
	if val, ok := templateMap[viewsPath]; ok {
		return val
	}
	templateMap[viewsPath] = template.Must(template.ParseFiles(views...))
	return templateMap[viewsPath]
}

func NewNavBuilder() NavBuilder {
	return NavBuilder{elements: map[string]interface{}{}}
}

type NavBuilder struct {
	elements map[string]interface{}
}

func (nb *NavBuilder) AddElement(title string, link string, icon string) {
	elm := map[string]interface{}{}
	elm["href"] = link
	elm["icon"] = icon
	elm["active"] = false
	nb.elements[title] = elm
}

func (nb *NavBuilder) GetNavigation(activeTitle string) map[string]interface{} {
	nav := map[string]interface{}{}
	for key, element := range nb.elements {
		elm := map[string]interface{}{}
		for propKey, propValue := range element.(map[string]interface{}) {
			elm[propKey] = propValue
		}
		if key == activeTitle {
			elm["active"] = true
		} else {
			elm["active"] = false
		}
		nav[key] = elm
	}
	return nav
}
