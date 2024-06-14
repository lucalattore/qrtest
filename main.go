package main

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	var wait time.Duration
	var port int64

	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Int64Var(&port, "port", 3000, "listening port")
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/", showParams)

	h2s := &http2.Server{}
	server := &http.Server{
		Addr:         "0.0.0.0:" + strconv.FormatInt(port, 10),
		Handler:      h2c.NewHandler(router, h2s),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Println("Listening on port", port)
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("Shutting down")
	os.Exit(0)
}

type ParamData struct {
	Params []Param
}

type Param struct {
	Name  string
	Value string
}

func showParams(w http.ResponseWriter, r *http.Request) {
	data := ParamData{}
	values := r.URL.Query()
	for k, v := range values {
		data.Params = append(data.Params, Param{Name: k, Value: strings.Join(v, ",")})
	}
	homeTemplate.Execute(w, data)
}

var homeTemplate = template.Must(template.New("").Parse(`
<html>
	<head>
    	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    	<meta name="color-scheme" content="light dark">
    	<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH" crossorigin="anonymous">
	</head>
	<body>
    	<div class="container-fluid p-5">
            {{range .Params}}
        	<div class="row col border rounded-1">
            	<div class="col-2">
					<div id="lbl" style="font-size: 20px;"><b>{{.Name}}</b></div>
            	</div>
            	<div class="col-10">
                	<div id="val" style="font-size: 20px;">{{.Value}}</div>
            	</div>
        	</div>
            {{end}}
    	</div>
	</body>
</html>
`))
