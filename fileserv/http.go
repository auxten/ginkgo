package fileserv

import (
	"github.com/labstack/echo/v4"
)

func ServFiles(e *echo.Echo, root string) {
	e.Static(root, "f")
	return
}
