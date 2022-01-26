module github.com/cmars/oniongrok

go 1.16

require (
	berty.tech/go-libtor v1.0.385
	github.com/cretz/bine v0.2.0
	github.com/frankban/quicktest v1.14.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
)

// Pending https://github.com/cretz/bine/pull/60
replace github.com/cretz/bine => ../bine
