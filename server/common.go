package server

import (
	"strings"

	"github.com/labstack/echo"
)

func AbsoluteURL(c echo.Context, resource string) string {
	if len(resource) > 0 && resource[0] != '/' {
		resource = "/" + resource
	}
	scheme := c.Scheme()
	if c.Request().Header.Get(echo.HeaderXForwardedProto) == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request().Host + resource
}

func SubdomainURL(c echo.Context, subdomain string) string {
	scheme := c.Scheme()
	if c.Request().Header.Get(echo.HeaderXForwardedProto) == "https" {
		scheme = "https"
	}
	locale := LocaleForRequest(c.Request())
	host := strings.TrimPrefix(c.Request().Host, "www.")
	host = strings.TrimPrefix(host, locale.Subdomain+".")
	if subdomain != "" {
		host = subdomain + "." + host
	}
	return scheme + "://" + host
}
