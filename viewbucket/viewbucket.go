package viewbucket

import (
	"html/template"
	"net/http"
	"strconv"

	"gopkg.in/mgutz/dat.v1"

	"fmt"

	"strings"

	"github.com/jaybeecave/base/datastore"
	"github.com/jaybeecave/base/flash"
	"github.com/jaybeecave/render"
)

type viewBucket struct {
	renderer *render.Render
	store    *datastore.Datastore
	w        http.ResponseWriter
	req      *http.Request
	Data     map[string]interface{}
}

func New(w http.ResponseWriter, req *http.Request, renderer *render.Render, store *datastore.Datastore) *viewBucket {
	viewBag := viewBucket{}
	viewBag.renderer = renderer
	viewBag.w = w
	viewBag.req = req
	viewBag.store = store
	viewBag.Data = make(map[string]interface{})
	for name, val := range store.ViewGlobals {
		viewBag.Add(name, val)
	}
	return &viewBag
}

func (viewBag *viewBucket) Add(key string, value interface{}) {
	viewBag.Data[key] = value
	// spew.Dump(viewBag.data)
}

func (viewBag *viewBucket) LoadNavItems() {
	var navItems []*NavItem
	err := viewBag.store.DB.
		Select("title", "slug").
		From("pages").
		QueryStructs(&navItems)
	if err != nil {
		panic(err)
	}
	viewBag.Add("NavItems", navItems)
}

func (viewBag *viewBucket) Render(status int, templateName string) {

	// automatically show the flash message if it exists
	msg, _ := flash.GetFlash(viewBag.w, viewBag.req, "InfoMessage")
	viewBag.Add("InfoMessage", msg) // if its blank it can be blank but atleast it will exist

	viewBag.renderer.HTML(viewBag.w, status, templateName, viewBag.Data)
}

var TemplateFunctions = template.FuncMap{
	"javascript": javascriptTag,
	"stylesheet": stylesheetTag,
	"image":      imageTag,
	"imagepath":  imagePath,
	"content":    content,
	"htmlblock":  htmlblock,
	"navigation": navigation,
	"link":       link,
	"title":      title,
}

func content(contents ...string) template.HTML {
	var str string
	for _, content := range contents {
		str += "<div class='standard'>" + content + "</standard>"
	}
	return template.HTML(str)
}

func javascriptTag(names ...string) template.HTML {
	var str string
	for _, name := range names {
		str += "<script src='/js/" + name + ".js' type='text/javascript'></script>"
	}
	return template.HTML(str)
}

func stylesheetTag(names ...string) template.HTML {
	var str string
	for _, name := range names {
		str += "<link rel='stylesheet' href='/css/" + name + ".css' type='text/css' media='screen'  />\n"
	}
	return template.HTML(str)
}

func imagePath(name string) string {
	return "/images/" + name
}

func imageTag(name string, class string) template.HTML {
	return template.HTML("<image src='" + imagePath(name) + "' class='" + class + "' />")
}

func htmlblock(page *Page, code string) template.HTML {
	html := "<div class='textblock editable' "
	html += " data-textblock='page-" + strconv.FormatInt(page.PageID, 10) + "-" + code + "'"
	html += " data-placeholder='#{placeholder}'> "
	html += getHTMLFromTextblock(page, code)
	html += "</div>"
	return template.HTML(html)
}

func link(text string, link string, viewBag *viewBucket) template.HTML {
	class := "link link-" + strings.ToLower(text)
	if strings.ToLower(link) == viewBag.req.URL.Path {
		class += " active"
	}
	return template.HTML(fmt.Sprintf(`<a class="%v" href="%v">%v</a>`, class, link, text))
}

func title(text string) string {
	return strings.Title(text)
}

func navigation(viewBag *viewBucket) template.HTML {
	html := ""
	if viewBag.Data["NavItems"] != nil {
		navItems := viewBag.Data["NavItems"].([]*NavItem)
		html = "<nav class='main-nav closed'>"
		for _, navItem := range navItems {
			html += "<a href='/" + navItem.Slug + "'>" + navItem.Title + "</a>"
		}
		html += "</nav>"
	}
	return template.HTML(html)
}

type Page struct {
	PageID     int64        `db:"page_id"`
	Title      string       `db:"title"`
	Body       string       `db:"body"`
	Slug       string       `db:"slug"`
	Template   string       `db:"template"`
	CreatedAt  dat.NullTime `db:"created_at"`
	UpdatedAt  dat.NullTime `db:"updated_at"`
	Textblocks []*Textblock
}

type NavItem struct {
	Title string `db:"title"`
	Slug  string `db:"slug"`
}

func (navItem *NavItem) getURL() string {
	return ""
}

type Textblock struct {
	TextblockID int64        `db:"textblock_id"`
	Code        string       `db:"code"`
	Body        string       `db:"body"`
	CreatedAt   dat.NullTime `db:"created_at"`
	UpdatedAt   dat.NullTime `db:"updated_at"`
	PageID      int64        `db:"page_id"`
}

func getHTMLFromTextblock(page *Page, code string) string {
	var body string
	for _, tb := range page.Textblocks {
		if tb.Code == code {
			body = tb.Body
		}
	}
	return body
}
