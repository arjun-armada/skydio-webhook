package mux

import (
	"net/http"

	"github.com/arjun-armada/skydio-webhook/apis/services/web-hook/route/checkapi"
)

func WebAPI() http.Handler {

	mux := http.NewServeMux()

	checkapi.Routes(mux)

	return mux
}
