# Film library

## Task
<p><p>It is necessary to develop the backend of the Filmoteka application, which provides a REST API for managing the movie database.</p>

<p>The application must support the following functions:</p>

<ul>
	<li>adding information about the actor (name, gender, date of birth),</li>
	<li>changing information about the actor.</li>
</ul>

<p>It is possible to change any information about the actor, both partially and completely:</p>

<ul>
	<li>deleting information about the actor,</li>
	<li>adding information about the movie.</li>
</ul>

<p>When adding a film, its title (at least 1 and no more than 150 characters), description (no more than 1000 characters), release date, rating (from 0 to 10) and a list of actors are indicated:</p>

<ul>
	<li>changing information about the movie.</li>
</ul>

<p>It is possible to change any information about the film, both partially and completely:</p>

<ul>
	<li>deleting information about the movie,</li>
	<li>getting a list of movies with the ability to sort by title, by rating, by release date. By default, sorting by rating (descending) is used,</li>
	<li>search for a movie by a fragment of the title, by a fragment of the actor's name,</li>
	<li>getting a list of actors, for each actor a list of films with his participation is also given,</li>
	<li>The API must be closed by authorization,</li>
	<li>two user roles are supported - regular user and administrator. An ordinary user has access only to receive data and search, an administrator has access to all actions. To simplify it, we can assume that the correspondence of users and roles is set manually (for example, directly through the database).</li>
</ul>

<p>Implementation requirements:</p>

<ul>
	<li>the implementation language is go,</li>
	<li>a relational database management system (preferably PostgreSQL) is used for data storage,</li>
	<li>API specification is provided (in Swagger 2.0 or OpenAPI 3.0 format).</li>
</ul>

<p>Bonus: the api-first (code generation from the specification) or code-first (specification generation from the code) approach is used.</p>

<ul>
	<li>To implement an http server, it is allowed to use only the standard http library (without frameworks),</li>
	<li>logging - the log should contain basic information about the requests being processed, errors,</li>
	<li>the application code is covered by unit tests by at least 70%,</li>
	<li>Dockerfile for building the image,</li>
	<li>docker-compose file for launching an environment with a running application and DBMS.</li>
</ul></p>

## Necessary tools

- Docker
- Go 1.17
- PostgreSQL

## Installation and launch

Create a file `.env` in the root of the project and specify the environment variables in it:
```
POSTGRES_USER=filmotheka_user
POSTGRES_PASSWORD=filmotheka_pass
POSTGRES_DB=filmotheka_db
POSTGRES_PORT=5432
POSTGRES_HOST=db
SERVER_PORT=8080
```
After that, you can launch the application using the command:
```bash
docker-compose up --build
```

## Testing

```bash
go test ./... -v -coverprofile=coverage/cover.out && go tool cover -html=coverage/cover.out -o coverage/cover.html && open coverage/cover.html
```

## Swagger

```bash
swag init -g cmd/filmotheka/main.go
```

## Documentation

Documentation is available at http://localhost:8080/swagger/index.html