module github.com/RobustRoundRobin/RRR

go 1.14

// Use replace to pick the correct 'fork' of go-ethereum see
// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
//
// This 'replace' lets us import quorum in a go.mod world *without* breaking
// quorums imports - quorum's imports are also 'replaced' by this directive.
// quorum's code base uses symlink trickery to do this because they don't use
// go modules yet.
replace github.com/ethereum/go-ethereum => github.com/jpmorganchase/quorum v2.6.0+incompatible

// We also need to be very picky about the goleveldb version for quorum's
// benefit. The go modules minimum version selection is missing the
// "DisableSeekCompaction" option
replace github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d

require (
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/aristanetworks/goarista v0.0.0-20210204223745-64c208b2a430 // indirect
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/cespare/cp v1.1.1 // indirect
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/elastic/gosigar v0.14.0 // indirect
	github.com/ethereum/go-ethereum v0.0.0-00010101000000-000000000000
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20191108122812-4678299bea08 // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-hclog v0.15.0 // indirect
	github.com/hashicorp/go-plugin v1.4.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/huin/goupnp v1.0.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jpmorganchase/quorum-hello-world-plugin-sdk-go v0.0.0-20201012205023-ad2eaa9cc244 // indirect
	github.com/karalabe/usb v0.0.0-20191104083709-911d15fe12a9 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.1.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/status-im/keycard-go v0.0.0-20200402102358-957c09536969 // indirect
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tv42/httpunix v0.0.0-20191220191345-2ba4b9c3382c // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/wsddn/go-ecdh v0.0.0-20161211032359-48726bab9208 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	google.golang.org/grpc v1.35.0 // indirect
	gopkg.in/karalabe/cookiejar.v2 v2.0.0-20150724131613-8dcd6a7f4951 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/olebedev/go-duktape.v3 v3.0.0-20200619000410-60c24ae608a6 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
)
