
# Chirpy Endpoints

## `/app/`

web site home page


## `/app/assets/`

location for website assets


## `GET /api/healthz`

The status of the server


## `POST /api/users`

Creation of a user.

Request Body:
``` json
{
	"email": "user@example.com",
	"password": "so and so password"
}
```

Response Body:

``` json
{
	"id": UUID,
	"created_at": TIMESTAMP,
	"updated_at": TIMESTAMP,
	"is_chirpy_red", BOOL,
	"email": users email
}
```

## `POST /api/users`

Update user information.

Request Body:
``` json
{
	"email": NEW USER EMAIL,
	"password": NEW USER PASSWORD
}
```

Response Body:
``` json
{
	"id": UUID,
	"created_at": TIMESTAMP,
	"updated_at": TIMESTAMP,
	"is_chirpy_red", BOOL,
	"email": users email
}
```


## `POST /api/login`

Login as user.

Request Body:
``` json
{
	"email": USER'S EMAIL,
	"password": USER'S PASSWORD
}
```

Response Body
``` json
{
	"id": USER'S UUID,
	"created_at": TIMESTAMP,
	"updated_at": TIMESTAMP,
	"email": USER'S EMAIL,
	"token": USER'S JSON WEB TOKEN,
	"refresh_token": TOKEN,
	"is_chirpy_red": BOOL
}
```


## `POST /api/refresh`

Refresh token for user.

Set the Authorization header as the refresh token for the request.

Response Body:
``` json
{
	"token": NEW JWT TOKEN
}
```


## `POST /api/revoke`

Revoke user's token.

Set the Authorization header as the refresh token for the request.

Response status: 204 No Content


## `POST /api/chirps` 

Create chirp for user.

Request Body:

``` json
{
	"body": CHIRP BODY,
	"user_id": UUID
}
```

As well needing the JWT of the user in the authorization header.

Response Body:
``` json
{
	"id": CHIRP ID,
	"user_id": UUID,
	"created_at": TIMESTAMP,
	"updated_at": TIMESTAMP,
	"body": CHRIP BODY
}
```


## `GET /api/chirps`

Get every chirp, or user chirps.

URL queries:

- `sort` sort by `asc` (the default) or desc
- `author_id` get user's chirps with user UUID


Response Body:
``` json
[
	{
		"id": CHIRP ID,
		"user_id": UUID,
		"created_at": TIMESTAMP,
		"updated_at": TIMESTAMP,
		"body": CHRIP BODY
	},
...
]
```

## `GET /api/chirps/{chirp_id}`

Get chirp by `chirp_id`.

Response Body:
``` json
{
	"id": CHIRP ID,
	"user_id": UUID,
	"created_at": TIMESTAMP,
	"updated_at": TIMESTAMP,
	"body": CHRIP BODY
}
```


## `DELETE /api/chirps/{chirp_id}`

Delete chirp by user.

Set authorization header to the JWT.

Response status as 204 No Content


## `GET /admin/metrics`

Get the stats of requests

**PLATFORM SET TO "dev"**

## `POST /admin/reset`

Reset status of request

**PLATFORM SET TO "dev"**


