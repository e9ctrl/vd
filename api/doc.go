// api package provides HTTP interface to expose simulator parameters and to control them via simple REST API.
// It contains methods used by HTTP REST server to handle queries and HTTP client functions to generate those queries.
// Client methods are used in CLI to provide a way to control parameters outside of TCP server.
// HTTP is launched together with TCP simulator server.
package api
