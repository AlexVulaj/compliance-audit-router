/*
Copyright © 2021 Red Hat, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package listeners

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"

	"github.com/openshift/compliance-audit-router/pkg/helpers"
	"github.com/openshift/compliance-audit-router/pkg/splunk"
)

type Listener struct {
	Path        string
	Methods     []string
	HandlerFunc gin.HandlerFunc
}

var Listeners = []Listener{
	{
		Path:        "/readyz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: RespondOKHandler,
	},
	{
		Path:        "/healthz",
		Methods:     []string{http.MethodGet},
		HandlerFunc: RespondOKHandler,
	},
	{
		Path:        "/api/v1/alert",
		Methods:     []string{http.MethodPost},
		HandlerFunc: ProcessAlertHandler,
	},
}

func InitRoutes(router *gin.Engine) {
	for _, listener := range Listeners {
		for _, method := range listener.Methods {
			router.Handle(method, listener.Path, listener.HandlerFunc)
		}
	}
}

// RespondOKHandler replies with a 200 OK and "OK" text to any request, for health checks
func RespondOKHandler(context *gin.Context) {
	context.Data(http.StatusOK, "text/plain", []byte("OK"))
}

func ProcessAlertHandler(context *gin.Context) {
	var alert splunk.Webhook

	if err := context.BindJSON(&alert); err != nil {
		var mr *helpers.MalformedRequest
		if errors.As(err, &mr) {
			context.JSON(mr.Status, mr.Msg)
		} else {
			log.Println(err.Error())
			context.Status(http.StatusInternalServerError)
		}
	}

	log.Println("Received alert from Splunk:", alert.Sid, alert.Result.Raw)
	fmt.Printf("%+v\n", alert)

	os.Exit(1)
}
