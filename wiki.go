package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/sjsafranek/DiffStore"
)

var (
	TEMPLATES *template.Template = template.Must(template.ParseFiles("templates/view.html"))
)

type WikiEngine struct {
	db Database
}

func (self *WikiEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.ViewHandler(w, r)
}

func (self *WikiEngine) getPageAtVersion(page string, version int) (*Page, error) {
	var ddata diffstore.DiffStore
	raw, err := self.db.Get("pages", page)
	if nil != err {
		logger.Warn(err)
		logger.Debug(page)
		ddata = diffstore.New()
	} else {
		ddata.Decode([]byte(raw))
	}

	if ddata.Length() <= version || -1 == version {
		logger.Debugf("Fetching current state: %v", page)
		return &Page{
			Title:           page,
			Data:            ddata.CurrentValue,
			CurrentVersion:  ddata.Length(),
			SelectedVersion: ddata.Length(),
		}, nil
	}

	logger.Debugf("Fetching previous state: %v", page)
	previous, err := ddata.GetPreviousByIndex(version)
	return &Page{
		Title:           page,
		Data:            previous,
		CurrentVersion:  ddata.Length(),
		SelectedVersion: version,
	}, err
}

func (self *WikiEngine) ViewHandler(w http.ResponseWriter, r *http.Request) {
	// redirect to random hash
	if "/" == r.URL.Path {
		seed := RandomStringOfLength(20)
		// 302 http status prevents browser from caching
		// and serving same results.
		http.Redirect(w, r, fmt.Sprintf("/%s", seed), http.StatusFound)
		return
	}

	switch {
	case "GET" == r.Method:
		self.httpGETHandler(w, r)
	case "POST" == r.Method:
		self.httpPOSTHandler(w, r)
	case "DELETE" == r.Method:
		self.httpDELETEHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (self *WikiEngine) getPageNameFromRequest(r *http.Request) string {
	page := r.URL.Path[1:]
	if len(page) == 0 {
		page = "index"
	}
	return page
}

func (self *WikiEngine) writeResponse(w http.ResponseWriter, status_code int, response string) {
	w.WriteHeader(status_code)
	fmt.Fprintln(w, response)
}

func (self *WikiEngine) writeApiErrorResponse(w http.ResponseWriter, status_code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	self.writeResponse(w, status_code, fmt.Sprintf(`{"status":"error", "message": "%v"}`, err.Error()))
}

func (self *WikiEngine) writeApiSuccessResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	self.writeResponse(w, http.StatusOK, `{"status":"ok"}`)
}

func (self *WikiEngine) httpGETHandler(w http.ResponseWriter, r *http.Request) {
	page := self.getPageNameFromRequest(r)

	version := -1
	version_param := r.URL.Query().Get("version")
	if "" != version_param {
		i, err := strconv.ParseInt(version_param, 10, 64)
		if nil != err {
			self.writeApiErrorResponse(w, http.StatusBadRequest, errors.New("version must be integer"))
			return
		}
		version = int(i)
	}

	p, err := self.getPageAtVersion(page, version)
	if nil != err {
		self.writeApiErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	mode := r.URL.Query().Get("mode")

	switch {

	case "edit" == mode:
		if nil != err {
			self.renderTemplate(w, "view", &Page{Title: page})
			return
		}

		self.renderTemplate(w, "view", p)

	case "" == mode:
		fallthrough

	case "view" == mode:
		data, err := p.Unmarshal()
		if nil != err {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, data)

	default:
		self.writeApiErrorResponse(w, http.StatusBadRequest, errors.New("Unsupported mode type"))
	}
}

func (self *WikiEngine) httpPOSTHandler(w http.ResponseWriter, r *http.Request) {
	page := self.getPageNameFromRequest(r)

	err := self.savePage(page, r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	self.writeApiSuccessResponse(w)
}

func (self *WikiEngine) httpDELETEHandler(w http.ResponseWriter, r *http.Request) {
	page := self.getPageNameFromRequest(r)

	err := self.db.Remove("pages", page)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	self.writeApiSuccessResponse(w)
}

func (self *WikiEngine) savePage(page string, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var ddata diffstore.DiffStore
	raw, err := self.db.Get("pages", page)
	if nil != err {
		logger.Warn(err)
		ddata = diffstore.New()
	} else {
		ddata.Decode([]byte(raw))
	}

	ddata.Update(string(body))
	enc, err := ddata.Encode()
	if nil != err {
		return err
	}

	return self.db.Set("pages", page, string(enc))
}

func (self *WikiEngine) renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := TEMPLATES.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
