package server

import (
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
