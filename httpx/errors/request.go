package errors

import (
	"net/http"
	"net/url"
)

func safeCloneRequest(req *http.Request) *http.Request {
	if req == nil {
		return nil
	}

	return &http.Request{
		Method:     req.Method,
		URL:        safeCloneURL(req.URL),
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Header:     *safeCloneHeader(&req.Header),
		// Body may have sensitive information,
		// besides all the data should be already parsed as Form and/or PostForm
		Body:             nil,
		ContentLength:    req.ContentLength,
		TransferEncoding: copyStringArray(req.TransferEncoding),
		Close:            req.Close,
		Host:             req.Host,
		Form:             *safeCloneForm(&req.Form),
		PostForm:         *safeCloneForm(&req.PostForm),
		// MultipartForm may have sensitive information
		MultipartForm: nil,
		// Trailer isn't that important for reporting purposes
		Trailer:    nil,
		RemoteAddr: req.RemoteAddr,
		RequestURI: req.RequestURI,
	}
}

func safeCloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	return &url.URL{
		Scheme: u.Scheme,
		// req.User may have sensitive information, like username and password
		Host:       u.Host,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: u.ForceQuery,
		RawQuery:   u.RawQuery,
		Fragment:   u.Fragment,
	}
}

var sensitiveHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
}

func safeCloneHeader(header *http.Header) *http.Header {
	if header == nil {
		return nil
	}
	safeHeader := http.Header{}
	for key, valueArray := range *header {
		if _, ok := sensitiveHeaders[key]; ok {
			continue
		}
		safeHeader[key] = copyStringArray(valueArray)
	}
	return &safeHeader
}

func copyStringArray(values []string) []string {
	if values == nil {
		return nil
	}
	safeArray := make([]string, len(values))
	copy(safeArray, values)
	return safeArray
}

var sensitiveFormKeys = map[string]bool{
	"password": true,
}

func safeCloneForm(form *url.Values) *url.Values {
	if form == nil {
		return nil
	}
	safeForm := url.Values{}
	for key, values := range *form {
		if _, ok := sensitiveFormKeys[key]; ok {
			continue
		}
		safeForm[key] = copyStringArray(values)
	}
	return &safeForm
}
