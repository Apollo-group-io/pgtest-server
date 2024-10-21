module pgtestserver

go 1.23.2

require github.com/rubenv/pgtest v1.1.0

require github.com/lib/pq v1.3.0 // indirect

replace github.com/rubenv/pgtest => github.com/Apollo-group-io/pgtest v1.2.0
