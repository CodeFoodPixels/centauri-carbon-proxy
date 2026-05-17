package types

import "net/http"

type RouteMap map[string]func(http.ResponseWriter, *http.Request)
