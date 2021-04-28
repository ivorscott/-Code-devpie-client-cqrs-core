package handlers

import (
	"github.com/devpies/devpie-client-core/projects/internal/mid"
	"github.com/devpies/devpie-client-core/projects/internal/platform/auth0"
	"github.com/devpies/devpie-client-core/projects/internal/platform/database"
	"github.com/devpies/devpie-client-core/projects/internal/platform/web"
	"log"
	"net/http"
	"os"
)

func API(shutdown chan os.Signal, repo *database.Repository, log *log.Logger, origins string,
	auth0Audience, auth0Domain, auth0MAPIAudience, auth0M2MClient, auth0M2MSecret string) http.Handler {

	a0 := &auth0.Auth0{
		Repo:         repo,
		Domain:       auth0Domain,
		Audience:     auth0Audience,
		M2MSecret:    auth0M2MSecret,
		M2MClient:    auth0M2MClient,
		MAPIAudience: auth0MAPIAudience,
	}

	app := web.NewApp(shutdown, log, mid.Logger(log), a0.Authenticate(), mid.Errors(log), mid.Panics(log))

	h := HealthCheck{repo: repo}

	app.Handle(http.MethodGet, "/api/v1/health", h.Health)

	t := Tasks{repo: repo, log: log, auth0: a0}
	c := Columns{repo: repo, log: log, auth0: a0}
	p := Projects{repo: repo, log: log, auth0: a0}

	app.Handle(http.MethodGet, "/api/v1/projects", p.List)
	app.Handle(http.MethodPost, "/api/v1/projects", p.Create)
	app.Handle(http.MethodGet, "/api/v1/projects/{pid}", p.Retrieve)
	app.Handle(http.MethodPut, "/api/v1/projects/{pid}", p.Update)
	app.Handle(http.MethodDelete, "/api/v1/projects/{pid}", p.Delete)
	app.Handle(http.MethodGet, "/api/v1/projects/{pid}/columns", c.List)
	app.Handle(http.MethodGet, "/api/v1/projects/{pid}/tasks", t.List)
	app.Handle(http.MethodPost, "/api/v1/projects/{pid}/columns/{cid}/tasks", t.Create)
	app.Handle(http.MethodPatch, "/api/v1/projects/tasks/{tid}", t.Update)
	app.Handle(http.MethodPatch, "/api/v1/projects/tasks/{tid}/move", t.Move)
	app.Handle(http.MethodDelete, "/api/v1/projects/columns/{cid}/tasks/{tid}", t.Delete)

	return Cors(origins).Handler(app)
}
