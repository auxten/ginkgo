package fileserv

import (
	"github.com/labstack/echo/v4"
)

func ServFiles(e *echo.Echo, root string, addr string) {

	e.Static(root, "f")

	e.Logger.Fatal(e.Start(addr))
}
