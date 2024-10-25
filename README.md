# Terraform Provider Wiki.js

Please use the [Documentation](https://registry.terraform.io/providers/tyclipso/wikijs/latest/docs) in the [Terraform registry](https://registry.terraform.io/providers/tyclipso/wikijs/latest)

This provider is a fork from the internal provider of [Startnext GmbH](https://www.startnext.com).
It implements more of the API components of Wiki.js and improves on documentation.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4
- [Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider locally

To distribute the provider locally, download the repository and run the following commands.

```shell
export PROVIDER_VERSION=1.0.0
export BINARY_ARCH=linux_amd64
CGO_ENABLED=0 go build -o "~/.terraform.d/plugins/terraform.local/tyclipso/wikijs/${PROVIDER_VERSION}/${BINARY_ARCH}/terraform-provider-wikijs_v${PROVIDER_VERSION}" -ldflags="-X 'main.Version=${PROVIDER_VERSION}'" main.go
```

Make sure to use the correct binary architecture.
You can query the values with `go env GOOS GOARCH`.
We commonly use one of the following:

- `linux_amd64`
- `darwin_arm64`
- `darwin_amd64`

The following section is an implied local provider config.
If you compile/intall the provider in a non implied location you need to provide the `~/.terraformrc` file.
Fill in the `path` under `filesystem_mirror` with your location.

```terraform
provider_installation {
  filesystem_mirror {
    path    = "/path/to/the/install/.terraform.d/plugins"
  }
  direct {
    exclude = ["terraform.local/*/*"]
  }
}
```

To use the provider you need to add this to your `required_providers`.

```terraform
wikijs = {
  source = "terraform.local/tyclipso/wikijs"
  version = "~> 1"
}
```

Replacing the provider may be beneficial.
You can achieve this with `terraform state replace-provider tyclipso/wikijs terraform.local/tyclipso/wikijs`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

To change graphql queries, edit `wikijs/genqclient.grapqhl` and run `go generate ./wikijs`

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
