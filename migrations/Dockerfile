FROM gomicro/goose

WORKDIR /migrations

COPY ./migrations/*.sql ./

CMD goose postgres "host=$POSTGRES_HOST user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB port=$POSTGRES_PORT sslmode=disable" up
