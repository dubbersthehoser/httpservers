
source .env

cmd="$1"

goose -dir sql/schema/ 'postgres' "$DB_URL" "$cmd"
