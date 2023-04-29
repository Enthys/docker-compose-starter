package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	//go:embed templates
	tmpFiles embed.FS

	//go:embed static
	staticFiles embed.FS
)

func main() {
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	d := NewDocker(workdir)
	if err = d.ReloadDockerCompose(); err != nil {
		log.Fatal(err)
	}
	log.Println(workdir)

	r := mux.NewRouter()

	staticFS := fs.FS(staticFiles)
	staticContent, err := fs.Sub(staticFS, "static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	r.PathPrefix("/reload").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := d.ReloadDockerCompose(); err != nil {
			log.Println(err)
			w.Write([]byte("Failed to reload" + err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.PathPrefix("/{composeId}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		composeId := vars["composeId"]
		if composeId == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found!"))
			return
		}

		tmp := template.Must(template.ParseFS(tmpFiles, "templates/compose.html"))
		composeFile := d.GetDockerCompose(composeId)
		if composeFile == nil {
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
			return
		}

		if err := tmp.Execute(w, composeFile); err != nil {
			panic(err)
		}
	})

	r.PathPrefix("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmp := template.Must(template.ParseFS(tmpFiles, "templates/index.html"))
		if err := tmp.Execute(w, struct {
			ComposeFiles []*DockerCompose
		}{ComposeFiles: d.AllDockerCompose()}); err != nil {
			panic(err)
		}
	})

	log.Fatal(http.ListenAndServe("127.0.0.1:8000", r))
	// currentWorkDir, err := os.Getwd()
	//
	//	if err != nil {
	//		log.Fatalf("Failed to get working directory: %s", err)
	//	}
	//
	// docker := NewDocker(currentWorkDir)
	//
	//	if err := docker.ReloadDockerCompose(); err != nil {
	//		log.Fatalf("Failed to load docker-compose files: %s", err)
	//	}
	//
	// someDC := docker.AllDockerCompose()[0]
	//
	//	if err := docker.StartDockerCompose(someDC); err != nil {
	//		log.Fatalf("Failed to start docker-compose file: %s", err)
	//	}
	//
	//	if err := docker.ListenToDockerCompose(someDC, func(line string) error {
	//		log.Println(line)
	//		return nil
	//	}); err != nil {
	//
	//		log.Fatalf("Something went wrong while listening in on docker-compose file: %s", err)
	//	}
}
