module pgtestserver

go 1.23.2

require (
	github.com/lib/pq v1.3.0 // indirect
	github.com/Apollo-group-io/pgtest v1.1.0
)

replace github.com/Apollo-group-io/pgtest => github.com/rubenv/pgtest v1.1.0