package handlers

import (
	"fmt"
	"github.com/devpies/devpie-client-core/users/api/publishers"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/devpies/devpie-client-core/users/domain/invites"
	"github.com/devpies/devpie-client-core/users/domain/memberships"
	"github.com/devpies/devpie-client-core/users/domain/projects"
	"github.com/devpies/devpie-client-core/users/domain/teams"
	"github.com/devpies/devpie-client-core/users/domain/users"
	"github.com/devpies/devpie-client-core/users/platform/auth0"
	"github.com/devpies/devpie-client-core/users/platform/database"
	"github.com/devpies/devpie-client-core/users/platform/web"
	"github.com/devpies/devpie-client-events/go/events"
)

type Team struct {
	repo        database.Storer
	log         *log.Logger
	auth0       auth0.Auther
	nats        *events.Client
	origins     string
	sendgridKey string
	query       TeamQueries
	publish     publishers.Publisher
}

type TeamQueries struct {
	team       teams.TeamQuerier
	project    projects.ProjectQuerier
	membership memberships.MembershipQuerier
	user       users.UserQuerier
	invite     invites.InviteQuerier
}

func (t *Team) Create(w http.ResponseWriter, r *http.Request) error {
	var nt teams.NewTeam
	var role memberships.Role = memberships.Administrator

	uid := t.auth0.UserByID(r.Context()) // mock

	if err := web.Decode(r, &nt); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if _, err := t.query.project.Retrieve(r.Context(), t.repo, nt.ProjectID); err != nil { //mock
		switch err {
		case projects.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case projects.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "creating team for project %q", nt.ProjectID)
		}
	}

	tm, err := t.query.team.Create(r.Context(), t.repo, nt, uid, time.Now()) // mock
	if err != nil {
		return err
	}

	nm := memberships.NewMembership{
		UserID: uid,
		TeamID: tm.ID,
		Role:   role.String(),
	}

	m, err := t.query.membership.Create(r.Context(), t.repo, nm, time.Now()) // mock
	if err != nil {
		return err
	}

	up := projects.UpdateProjectCopy{
		TeamID: &tm.ID,
	}

	if err = t.query.project.Update(r.Context(), t.repo, nt.ProjectID, up); err != nil {
		return err
	} // mock

	if t.nats != nil {
		err = t.publish.MembershipCreatedForProject(t.nats, m, nt.ProjectID, uid)
		if err != nil {
			return err
		}
	}

	return web.Respond(r.Context(), w, tm, http.StatusCreated)
}

func (t *Team) AssignExisting(w http.ResponseWriter, r *http.Request) error {
	tid := chi.URLParam(r, "tid")
	pid := chi.URLParam(r, "pid")
	uid := t.auth0.UserByID(r.Context())

	tm, err := t.query.team.Retrieve(r.Context(), t.repo, tid)
	if err != nil {
		return web.NewRequestError(err, http.StatusNotFound)
	}

	var up = projects.UpdateProjectCopy{
		TeamID:    &tm.ID,
		UpdatedAt: time.Now().UTC(),
	}

	err = t.query.project.Update(r.Context(), t.repo, pid, up)
	if err != nil {
		switch err {
		case projects.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case projects.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "updating project %q", pid)
		}
	}

	if t.nats != nil {
		err = t.publish.ProjectUpdated(t.nats, &tm.ID, pid, uid)
		if err != nil {
			return err
		}
	}
	return web.Respond(r.Context(), w, nil, http.StatusOK)
}

func (t *Team) LeaveTeam(w http.ResponseWriter, r *http.Request) error {
	tid := chi.URLParam(r, "tid")

	uid := t.auth0.UserByID(r.Context())

	mid, err := t.query.membership.Delete(r.Context(), t.repo, tid, uid)
	if err != nil {
		switch err {
		case teams.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case teams.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "looking for team %q", tid)
		}
	}

	if t.nats != nil {
		err = t.publish.MembershipDeleted(t.nats, mid, uid)
		if err != nil {
			return err
		}
	}
	return web.Respond(r.Context(), w, nil, http.StatusOK)
}

func (t *Team) Retrieve(w http.ResponseWriter, r *http.Request) error {
	tid := chi.URLParam(r, "tid")

	tm, err := t.query.team.Retrieve(r.Context(), t.repo, tid)
	if err != nil {
		switch err {
		case teams.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case teams.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "looking for team %q", tid)
		}
	}

	return web.Respond(r.Context(), w, tm, http.StatusOK)
}

func (t *Team) List(w http.ResponseWriter, r *http.Request) error {
	uid := t.auth0.UserByID(r.Context())

	tms, err := t.query.team.List(r.Context(), t.repo, uid)
	if err != nil {
		switch err {
		case teams.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case teams.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "looking for user's teams")
		}
	}

	return web.Respond(r.Context(), w, tms, http.StatusOK)
}

func (t *Team) CreateInvite(w http.ResponseWriter, r *http.Request) error {
	var list invites.NewList

	tid := chi.URLParam(r, "tid")
	link := strings.Split(t.origins, ",")[0]

	if err := web.Decode(r, &list); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	token, err := t.auth0.RetrieveToken()
	if err == auth0.ErrNotFound || t.auth0.IsExpired(token) {
		var nt auth0.NewToken
		var tk auth0.Token

		nt, err = t.auth0.NewManagementToken()
		if err != nil {
			return err
		}
		// clean table before persisting
		if err = t.auth0.DeleteToken(); err != nil {
			return err
		}

		tk, err = t.auth0.PersistToken(nt, time.Now())
		if err != nil {
			return err
		}
		token = tk
	}

	for _, email := range list.Emails {
		ni := invites.NewInvite{
			TeamID: tid,
		}
		// when user exists
		u, err := t.query.user.RetrieveByEmail(t.repo, email)
		if err != nil {
			var au auth0.AuthUser

			au, err = t.auth0.CreateUser(token, email)
			if err != nil {
				return err
			}

			nu := users.NewUser{
				Auth0ID:       au.Auth0ID,
				Email:         au.Email,
				EmailVerified: au.EmailVerified,
				FirstName:     au.FirstName,
				Picture:       au.Picture,
			}

			var us users.User

			us, err = t.query.user.Create(r.Context(), t.repo, nu, time.Now())
			if err != nil {
				return err
			}

			ni.UserID = us.ID

			if err = t.auth0.UpdateUserAppMetaData(token, au.Auth0ID, us.ID); err != nil {
				return err
			}

			link, err = t.auth0.ChangePasswordTicket(token, au, link)
			if err != nil {
				return err
			}

		} else {
			ni.UserID = u.ID
		}

		if err = t.SendMail(email, link); err != nil {
			return err
		}

		_, err = t.query.invite.Create(r.Context(), t.repo, ni, time.Now())
		if err != nil {
			return err
		}
	}

	return web.Respond(r.Context(), w, nil, http.StatusCreated)
}

func (t *Team) RetrieveInvites(w http.ResponseWriter, r *http.Request) error {
	uid := t.auth0.UserByID(r.Context())

	is, err := t.query.invite.RetrieveInvites(r.Context(), t.repo, uid)
	if err != nil {
		switch err {
		case teams.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case teams.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "searching team invites for %q", uid)
		}
	}

	var res []invites.InviteEnhanced
	for _, invite := range is {
		team, err := t.query.team.Retrieve(r.Context(), t.repo, invite.TeamID)
		if err != nil {
			return err
		}
		ie := invites.InviteEnhanced{
			ID:         invite.ID,
			UserID:     invite.UserID,
			TeamID:     invite.TeamID,
			TeamName:   team.Name,
			Read:       invite.Read,
			Accepted:   invite.Accepted,
			Expiration: invite.Expiration,
			UpdatedAt:  invite.UpdatedAt,
			CreatedAt:  invite.CreatedAt,
		}
		res = append(res, ie)
	}

	return web.Respond(r.Context(), w, res, http.StatusOK)
}

func (t *Team) UpdateInvite(w http.ResponseWriter, r *http.Request) error {
	var update invites.UpdateInvite
	var role memberships.Role = memberships.Editor

	uid := t.auth0.UserByID(r.Context())
	tid := chi.URLParam(r, "tid")
	iid := chi.URLParam(r, "iid")

	if err := web.Decode(r, &update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	iv, err := t.query.invite.Update(r.Context(), t.repo, update, uid, iid, time.Now())
	if err != nil {
		return err
	}

	if update.Accepted {
		nm := memberships.NewMembership{
			UserID: uid,
			TeamID: tid,
			Role:   role.String(),
		}
		m, err := t.query.membership.Create(r.Context(), t.repo, nm, time.Now())
		if err != nil {
			return err
		}

		if t.nats != nil {
			err = t.publish.MembershipCreated(t.nats, m, uid)
			if err != nil {
				return err
			}
		}
	}

	return web.Respond(r.Context(), w, iv, http.StatusOK)
}

func (t *Team) SendMail(email, link string) error {
	from := mail.NewEmail("DevPie", "people@devpie.io")
	subject := "You've been invited to a Team on DevPie!"
	to := mail.NewEmail("Invitee", email)

	html := ""
	html += "<strong>Join Devpie</strong>"
	html += "<br/>"
	html += "<p>To accept your invitation, <a href=\"%s\">create an account</a>.</p>"
	htmlContent := fmt.Sprintf(html, link)

	plainTextContent := fmt.Sprintf("You've been invited to a Team on DevPie! %s ", link)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(t.sendgridKey)

	response, err := client.Send(message)
	if err != nil {
		return err
	}

	t.log.Println(response.StatusCode)
	t.log.Println(response.Body)
	t.log.Println(response.Headers)

	return nil
}
