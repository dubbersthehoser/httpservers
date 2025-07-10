# Http Server

A Boot.dev lesson for creating a http server. An introduction to http server by creating monolith back-end twitter clown called Chirpy.

# The Chirpy Web Server Features

- User Creation
- User posts and post deletion. 
- Passwords are stored hashed.
- Login in with JWT (Json Web Token) with token refreshing.


# Build

## Creating `.env`

There're four variables to configure before running chirp.

1. `DB_URL` the Postgres url connection.
1. `PLATFORM` set to `"dev"` for enable dev requests, like reset to reset values in the database.
1. `JWT_SECRET_KEY` for the Json Web Token signing (**ONLY USING HMAC**)
1. `POLKA_KEY` a fake web hook serves key for a payed Chirpy Red serves.

## Postgres and Goose

Install Goose for database migrations 

- `go install github.com/pressly/goose/v3/cmd/goose@latest`

Once your Postgres database is up and its connection string is set in `.env`, run `./gooser.sh up` to migrated the database for server use.


