// Package alogin defines the Login interface for handling login in web
// applications.
//
// The implementations of Login should be used with the
// //infra-sk/modules/alogin-sk control.
package alogin

import (
	"net/http"

	"go.skia.org/infra/go/roles"
)

// EMail is an email address.
type EMail string

// String returns the email address as a string.
func (e EMail) String() string {
	return string(e)
}

// NotLoggedIn is the EMail value used to indicate a user is not logged in.
const NotLoggedIn EMail = ""

// Status describes the logged in status for a user. Email will be empty if the
// user is not logged in.
type Status struct {
	// LoginURL is the URL to visit to log in.
	LoginURL string `json:"login"`

	// LogoutURL is the URL to visit to log out.
	LogoutURL string `json:"logout"`

	// EMail is the email address of the logged in user, or the empty string if
	// they are not logged in.
	EMail EMail `json:"email"`

	// All the Roles of the current user.
	Roles roles.Roles `json:"roles"`
}

// Login provides information about the logged in status of http.Requests.
type Login interface {
	// LoggedInAs returns the email of the logged in user, or the empty string
	// of they are not logged in.
	LoggedInAs(r *http.Request) EMail

	// NeedsAuthentication will send the right response to the user if they
	// attempt to use a resource that requires authentication, such as
	// redirecting them to a login URL or returning an http.StatusForbidden
	// response code.
	//
	// TODO(jcgregorio) Can be removed once Perf migrates fully to auth-proxy.
	NeedsAuthentication(w http.ResponseWriter, r *http.Request)

	// Status returns the logged in status and other details about the current
	// user.
	Status(r *http.Request) Status

	// All the authorized Roles for a user.
	Roles(r *http.Request) roles.Roles

	// Returns true if the currently logged in user has the given Role.
	HasRole(r *http.Request, role roles.Role) bool
}
