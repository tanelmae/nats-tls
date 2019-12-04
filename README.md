# nats-tls
Tool for generating NATS TLS keys and certificates based on single yaml configuration file. There is also section for `default` settings that can be used for shared settings for key/cert pairs to reduce duplication. Mostly aimed at developer workflows.

Signature Algorithm: sha256WithRSAEncryption

Configuration:
```
# key can be default, ca or name for arbitary key/cert pair
# ca section is is required.
# To create a signed certificat at least something about it needs to be defined byt the specific cert. 
# Defining `subject.cn` in the examples is an arbitary choice for that purpose.
<key>:
  # path to dir where output will be stored	
  path: local
  # private key length
  key_length: 2048
  # X509v3 Subject Alternative Name DNS fields
  # Put the certificate domain name here if needed - when used by server
  dns:
  	- example.com
  # Ceritificate subject suppors just these 3 fields
  subject:
    cn: common-name.xyz
    org: "My Laptop"
    country: "Elbonia"
  # Certificate expiration period. Supported time units: years, months, days, hours, minutes, seconds
  ttl: 24 hours
```

Example:
```
default:

  path: local
  key_length: 2048
  subject:
    org: "My Laptop"
    country: "Elbonia"
  ttl: 24 hours
ca:
  # Uses different key length
  key_length: 4096
  subject:
    cn: "Local CA"
route:
  # domain specific to user of that certificate
  dns:
    - "my-server.io"
  subject:
    cn: "NATS route"
server:
  dns:
    - "localhost"
  subject:
    cn: "NATS server"
# Not client certificate created as it is commented out
# client:
#  subject:
#    cn: "NATS client"
```

### CLI
```
user@computer: $ nats-tls -h
Usage of nats-tls:
  -config string
    	Path to config file (default "nats-tls.yaml")
  -debug
    	Run in debug mode
  -v	Version
 
  ```


For examples check the files in `examples` directory.


#### Build locally:
```
go mod vendor
go build -mod=readonly -o nats-tls cmd/main.go
```

#### Release downloads:
https://github.com/tanelmae/nats-tls/releases

#### Install with Homebrew:

Binaries provided for Darwin_x86_64 and Linux_x86_64
```
brew install tanelmae/brew/nats-tls
```
