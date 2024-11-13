module github.com/refraction-networking/utls

go 1.22.0

retract (
	v1.4.1 // #218
	v1.4.0 // #218 panic on saveSessionTicket
)

require (
	github.com/andybalholm/brotli v1.1.0
	github.com/cloudflare/circl v1.5.0
	github.com/imroc/req/v3 v3.48.0
	github.com/klauspost/compress v1.17.9
	golang.org/x/crypto v0.27.0
	golang.org/x/net v0.29.0
	golang.org/x/sys v0.25.0
)

require golang.org/x/text v0.18.0 // indirect
