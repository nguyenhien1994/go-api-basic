package auth

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/rs/xid"

	"github.com/gilcrest/go-API-template/env"
	"github.com/gilcrest/go-API-template/errors"
)

// CreateClientRequest is used for the client service
type CreateClientRequest struct {
	ClientName        string `json:"client_name"`
	ClientHomeURL     string `json:"homepage_url"`
	ClientDescription string `json:"client_description"`
	RedirectURI       string `json:"redirect_uri"`
	Username          string `json:"username"`
}

// Client is used for the client service and response
type Client struct {
	id            string
	Name          string
	HomeURL       string
	Description   string
	RedirectURI   string
	PrimaryUserID string
	secret        string
	DMLTime       time.Time
}

// grant_types     VARCHAR(80),
// scope           VARCHAR(4000),

// ID is a getter for Client.id
func (c *Client) ID() string {
	return c.id
}

// SetID is a setter for Client.id
func (c *Client) SetID() {
	c.id = xid.New().String()
}

// Secret is a getter for Client.secret
func (c *Client) Secret() string {
	return c.secret
}

// SetSecret is a setter for Client.secret
func (c *Client) SetSecret() {
	c.secret = strings.Join([]string{xid.New().String(), xid.New().String()}, "-")
}

// NewClient is a constructor for the Client struct
func NewClient(ctx context.Context, env *env.Env, c *CreateClientRequest) (*Client, error) {
	const op errors.Op = "auth.CreateClient"

	client := new(Client)

	client.Name = c.ClientName
	client.HomeURL = c.ClientHomeURL
	client.Description = c.ClientDescription
	client.RedirectURI = c.RedirectURI
	client.PrimaryUserID = c.Username

	client.SetID()
	client.SetSecret()

	return client, nil
}

// CreateClientDB creates a client/app in the database
func (c *Client) CreateClientDB(ctx context.Context, tx *sql.Tx) (*sql.Tx, error) {
	const op errors.Op = "auth.CreateClientDB"

	var dmlTime time.Time

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `select demo.create_client (
		p_client_id => $1,
		p_client_name => $2,
		p_homepage_url => $3,
		p_app_description => $4,
		p_redirect_uri => $5,
		p_client_secret => $6,
		p_user_id => $7)`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		c.id,            //$1
		c.Name,          //$2
		c.HomeURL,       //$3
		c.Description,   //$4
		c.RedirectURI,   //$5
		c.secret,        //$6
		c.PrimaryUserID) //$7

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&dmlTime); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// set the DMLTime field to the create_date set as part of the insert in
	// the stored function call above
	c.DMLTime = dmlTime

	return tx, nil

}