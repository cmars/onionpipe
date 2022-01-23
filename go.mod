module github.com/cmars/oniongrok

go 1.16

require (
	github.com/cretz/bine v0.2.0
	github.com/frankban/quicktest v1.14.0
	github.com/ipsn/go-libtor v1.0.380
	github.com/urfave/cli/v2 v2.3.0
)

// Pending https://github.com/cretz/bine/pull/60
replace github.com/cretz/bine => github.com/cmars/bine v0.0.0-20220123065450-c094cc4fe044
