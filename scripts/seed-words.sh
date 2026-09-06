#!/bin/sh

set -e

DB="./internal/db/wordly.db"
WORDS="./scripts/words.json"

echo "Seeding Wordly words..."

{
    echo "BEGIN TRANSACTION;"

    jq -r '.words[]' "$WORDS" |
    while IFS= read -r word; do
        printf "INSERT OR IGNORE INTO words (word) VALUES ('%s');\n" "$word"
    done

    echo "COMMIT;"
} | sqlite3 "$DB"

COUNT=$(sqlite3 "$DB" "SELECT COUNT(*) FROM words;")

echo "Done. $COUNT words in database."
