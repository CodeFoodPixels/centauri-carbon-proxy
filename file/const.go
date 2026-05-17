package file

var ALLOWED_REQUEST_HEADERS = map[string][]string{
	"GET": {
		"accept",
		"accept-language",
		"accept-encoding",
		"priority",
		"user-agent",
		"range",
		"if-none-match",
		"if-modified-since",
	},
	"HEAD": {
		"accept",
		"accept-language",
		"accept-encoding",
		"priority",
		"user-agent",
		"range",
		"if-none-match",
		"if-modified-since",
	},
	"OPTIONS": {
		"origin",
		"access-control-request-method",
		"access-control-request-headers",
	},
	"POST": {
		"user-agent",
		"accept",
		"accept-language",
		"accept-encoding",
		"content-length",
		"content-type",
		"origin",
	},
	"WS": {
		"connection",
		"upgrade",
		"origin",
		"sec-websocket-extensions",
		"sec-websocket-key",
		"sec-websocket-protocol",
		"sec-websocket-version",
	},
}

var ALLOWED_RESPONSE_HEADERS = map[string][]string{
	"GET": {
		"content-length",
		"content-type",
		"content-encoding",
		"etag",
		"cache-control",
		"last-modified",
		"accept-ranges",
	},
	"HEAD": {
		"content-length",
		"content-type",
		"content-encoding",
		"etag",
		"cache-control",
		"last-modified",
		"accept-ranges",
	},
	"OPTIONS": {
		"access-control-allow-origin",
		"access-control-allow-methods",
		"access-control-allow-headers",
		"access-control-max-age",
		"content-length",
	},
	"POST": {
		"content-length",
		"content-type",
		"content-encoding",
	},
}

var TRANSFORMABLE_MIME_TYPES = []string{
	"text/plain",
	"text/css",
	"text/html",
	"text/javascript",
	"application/json",
}

var CACHEABLE_MIME_TYPES = []string{
	"text/plain",
	"text/css",
	"text/html",
	"text/javascript",
	"application/json",
	"image/apng",
	"image/avif",
	"image/gif",
	"image/jpeg",
	"image/png",
	"image/svg+xml",
	"image/webp",
}
