# url shortener backend

## stack

- go 1.22
- postgres
- goose

## build

In the project root directory there is a .env file that manage environment for local development.
In the ./docker directory there are docker-compose.yml and .env file for dev with docker.

- fill up ./docker/.env on the base of ./.env.template.

- run `docker-compose up --build` to build docker containers. Notice that db in postgres container won't have time to start before app at the first build, so you need to stop and re-up containers.

- stop (CTRL+C) and re-up containers with `docker-compose up`.
