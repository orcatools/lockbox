module github.com/orcatools/lockbox

go 1.14

replace golang.org/x/crypto => github.com/ProtonMail/crypto v0.0.0-20200416114516-1fa7f403fb9c

require (
	github.com/pquerna/otp v1.2.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
)
