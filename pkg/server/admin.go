package server

import (
	"errors"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) listShortlinks(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	pageParam := request.URL.Query().Get("page")
	if pageParam == "" {
		pageParam = "0"
	}

	page, err := strconv.ParseInt(pageParam, 10, 64)
	if err != nil {
		http.Error(writer, "Param page is not an integer", http.StatusBadRequest)
		return
	}

	size := int64(5)

	shortlinks, total, err := s.repo.GetEntries(request.Context(), page, size)
	if err != nil {
		http.Error(writer, "Error getting shortlinks", 500)
		return
	}

	csrfToken := generateCsrf(writer, request)

	err = templates["list.page.gohtml"].Execute(writer, listTemplateData{
		Shortlinks: shortlinks,
		Page:       page,
		Total:      total,
		Size:       size,
		CSRF:       csrfToken,
	})
	if err != nil {
		log.Errorw("error rendering page", "url", request.URL.String(), "error", err)
		http.Error(writer, "Error rendering page", 500)
	}
}

func (s *Server) editShortlink(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	code := params.ByName("code")

	url, err := s.repo.GetLinkForCode(request.Context(), code)
	if err != nil {
		http.Error(writer, "Error getting database data", 500)
		return
	}

	csrfToken := generateCsrf(writer, request)

	err = templates["edit.page.gohtml"].Execute(writer, editTemplateData{
		Code: code,
		URL:  url,
		CSRF: csrfToken,
	})
	if err != nil {
		log.Errorw("error rendering page", "url", request.URL.String(), "error", err)
		http.Error(writer, "Error rendering page", 500)
	}
}

func (s *Server) createOrEdit(writer http.ResponseWriter, request *http.Request, param httprouter.Params) {
	err := checkCsrf(request)
	if err != nil {
		http.Error(writer, "csrf token error", 403)
		return
	}

	formCode := request.Form.Get("code")
	if formCode == "" {
		http.Error(writer, "Missing code in form data", 400)
		return
	}

	formUrl := request.Form.Get("url")
	if formUrl == "" {
		http.Error(writer, "Missing url in form data", 400)
		return
	}

	existingCode := param.ByName("code")

	if existingCode != formCode && existingCode != "" {
		err = s.repo.DeleteCode(request.Context(), existingCode)
		if err != nil {
			log.Errorw("delete code error", "code", existingCode, "error", err)
			http.Error(writer, "Could not remove old code", 500)
			return
		}
	}

	err = s.repo.SetLinkForCode(request.Context(), formCode, formUrl)
	if err != nil {
		log.Errorw("set code error", "code", formCode, "url", formUrl, "error", err)
		http.Error(writer, "Could not save shortlink", 500)
		return
	}

	http.Redirect(writer, request, "/admin/shortlinks", 302)
	return
}

func (s *Server) deleteShortlink(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	err := checkCsrf(request)
	if err != nil {
		http.Error(writer, "csrf token error", 403)
		return
	}

	code := params.ByName("code")
	err = s.repo.DeleteCode(request.Context(), code)
	if err != nil {
		http.Error(writer, "could not delete shortlink", 500)
		return
	}

	http.Redirect(writer, request, "/admin/shortlinks", 302)
}

func generateCsrf(writer http.ResponseWriter, request *http.Request) string {
	tokenValue := uuid.New().String()
	if csrfCookie, err := request.Cookie("__Host-CSRF"); err == nil {
		tokenValue = csrfCookie.Value
	} else {
		http.SetCookie(writer, &http.Cookie{
			Name:     "__Host-CSRF",
			Value:    tokenValue,
			Path:     "/",
			Expires:  time.Now().Add(30 * time.Minute),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}
	return tokenValue
}

func checkCsrf(request *http.Request) error {
	csrfCookie, err := request.Cookie("__Host-CSRF")
	if err != nil {
		return errors.New("csrf cookie not set")
	}

	err = request.ParseForm()
	if err != nil {
		log.Warnw("form not parsable", "path", request.URL.Path, "error", err)
		return errors.New("form not parsable")
	}

	csrfFormValue := request.Form.Get("_csrf")
	if csrfCookie.Value != csrfFormValue {
		return errors.New("mismatched csrf token")
	}
	return nil
}
