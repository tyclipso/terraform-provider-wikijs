# Terraform Provider Wiki.js

Please use the [Documentation](https://registry.terraform.io/providers/tyclipso/wikijs/latest/docs) in the [Terraform registry](https://registry.terraform.io/providers/tyclipso/wikijs/latest)

This provider is a fork from the internal provider of [Startnext GmbH](https://www.startnext.com).
It implements more of the API components of Wiki.js and improves on documentation.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4
- [Go](https://golang.org/doc/install) >= 1.24

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using Gos `install` command:

```shell
go install
```

### `GNUmakefile`

Alternatively a `Makefile` is included that should cover nearly everything needed to interact with the provider repository.
Use `make help` to get a list with a brief description of each target.
The `make` default nearly acts like a `go install`.

## `.envrc`

Certain variables can be managed easier by using `direnv` with an `.envrc`.
The included `example.envrc` can be copied and customized as `.envrc`.
It currently handles go tool invocation as well as setting up build variables.
This allows for skipping all `export` sections in the build instructions as well as not having to call `go tool …`.

## Using the provider locally

There are multiple ways how to interact with the provider locally.

### `dev_overrides`

Terraform would normally record `zh:` for each built provider and write these to the lock file.
This hinders a rapid development cycle of building, testing, fixing because the binary is always different.

For this you can use `dev_overrides`.

Provide the following file as `.terraformrc` in your terraform directory where the provider should be used.

```terraform
provider_installation {
  dev_overrides {
    # Prevent terraform init to handle the given provider and point to where
    # the binary can be found
    "tyclipso/wikijs" = "../terraform-provider-wikijs/bin"
  }
  # load everything else as normal
  direct {}
}
```

> **WARNING**:
> `dev_overrides` accept relative paths with `.` or `..` but do not expand `~`, `$HOME` or similiar.

Now use `TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"` before terraform invocation or `export` it.

With `go install` or `make` you can build a new provider binary which is automatically picked up by terraform.

To undo the change unset the variable `TF_CLI_CONFIG_FILE`.

### `filesystem_mirror`

To use a pre-release version that works nearly the same as an official release you can use the `filesystem_mirror` option.
This can even be used in conjunction with `dev_overrides` and you only need to comment out the block and run a `terraform init -upgrade` to use the local version.

> **WARNING**:
> The local version string should be set to something that the official release will never have.
> Otherwise there is a good chance for version conflicts with the `filesystem_mirror` setting because it uses the official registry url.
> The `Makefile` sets it to `99.0.$(date +%s)` with the last being the unix timestamp making it easier to progress through versions if needed.

To build the pre-release version run the following commands.
Variables need to match the `provider_installation` block further down.

```shell
export PLUGIN_DIR="$HOME/.terraform.d/plugins"
export REGISTRY="registry.terraform.io"
export PROVIDER="tyclipso/wikijs"
export PROVIDER_VERSION="99.0.$(date +%s)"
export PLATFORM="$(go env GOOS)_$(go env GOARCH)"
export BINARY="terraform-provider-wikijs"
CGO_ENABLED=0 go build \
  -ldflags="-X 'main.Version=$PROVIDER_VERSION'" \
  -o "$PLUGIN_DIR/$REGISTRY/$PROVIDER/$PROVIDER_VERSION/$PLATFORM/$BINARY_v$PROVIDER_VERSION" .
```

For a new build rerun `export PROVIDER_VERSION="99.0.$(date +%s)"` before the build command.
This sets the version string to a new patch with the current unix timestamp.

The configuration file `.terraformrc` placed in the directory where the provider is to be used needs the following content.
Customize as necessary and use the variable contents of the build.

```terraform
provider_installation {
  # Uncomment to switch back to `dev_overrides` which superseed
  # `filesystem_mirror` setttings
  # dev_overrides {
  #   "tyclipso/wikijs" = "../terraform-provider-wikijs/bin"
  # }
  filesystem_mirror {
    # Path of your plugins directory
    path    = "/path/to/home/.terraform.d/plugins"
    # Use local system for the following provider
    include = ["registry.terraform.io/tyclipso/wikijs"]
  }
  direct {
    # Do not pull the following provider from remote registry 
    exclude = ["registry.terraform.io/tyclipso/wikijs"]
  }
}
```

> **WARNING**:
> `filesystem_mirror.path` accepts relative paths with `.` or `..` but do not expand `~`, `$HOME` or similiar.

Now use `TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"` before terraform invocation or `export` it.

Make sure that the `required_providers` block allows/requires a `v99.x.x` of the provider.
Either repin the provider or use `>=`.

To undo the change unset the variable `TF_CLI_CONFIG_FILE`.

### `terraform.local` registry

Terraform allows for different provider registries.
You can use the `filesystem_mirror` as a flat-file registry by populating the necessary files and directories.
The "url" or name is free to choose but to make it not interfere with other registries and still use a meaningful name choose `terraform.local`.
This can be used as the most "official distribution" way as it does not need special versions to prevent clashes, rely on `dev_overrides` or project-local `.terraformrc`.
You can even go through the list of `git tag`s and build each and every one of those to test versions.

> **CAVEAT**:
> This approach needs the most rewriting in your testbed as you need to replace the provider.

The config can be set for the whole machine in the default `~/.terraformrc`.
Customize as necessary and use the variable contents of the build.

```terraform
provider_installation {
  filesystem_mirror {
    # Path of your plugins directory
    path    = "/path/to/home/.terraform.d/plugins"
  }
  direct {
    # declare terraform.local as offline/non-direct registry
    exclude = ["terraform.local/*/*"]
  }
}
```

> **WARNING**:
> `filesystem_mirror.path` accepts relative paths with `.` or `..` but do not expand `~`, `$HOME` or similiar.

To build the version run the following.
If the `PROVIDER_VERSION` is non existent, skip the `git checkout "v$PROVIDER_VERSION"`.

```shell
export PLUGIN_DIR="$HOME/.terraform.d/plugins"
export REGISTRY="terraform.io/local"
export PROVIDER="tyclipso/wikijs"
export PROVIDER_VERSION="1.0.0"
export PLATFORM="$(go env GOOS)_$(go env GOARCH)"
export BINARY="terraform-provider-wikijs"
git checkout "v$PROVIDER_VERSION"
CGO_ENABLED=0 go build \
  -ldflags="-X 'main.Version=$PROVIDER_VERSION'" \
  -o "$PLUGIN_DIR/$REGISTRY/$PROVIDER/$PROVIDER_VERSION/$PLATFORM/$BINARY_v$PROVIDER_VERSION" .
```

Update all `required_provider` blocks where the provider is referenced.

```terraform
wikijs = {
  source = "terraform.local/tyclipso/wikijs"
  version = "~> 1"
}
```

Then you need to replace the provider for all existing resources.
If you start from a clean slate you can skip the following command
`terraform state replace-provider tyclipso/wikijs terraform.local/tyclipso/wikijs`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider for `dev_overrides`, run `make` or `make build`.
This will build the provider and put the binary in the `$GOPATH/bin` directory.
This is the path the project-local `.terraformrc` should point to.

To generate or update documentation, run `make docs`.

To change graphql queries, edit `wikijs/genqclient.grapqhl` and run `make graphql`.

To update both at once run `make generate`.

There are various other helpful `make` target.
Check `make help` or the `GNUmakefile` to evaluate what helps.

### Workaround needed

`make snapshot` uses the `.goreleaser.yml` which currently is setup for use in GitHub Actions.
Specifically the GPG signing fingerprint needs rework to accept a local override without compromising GitHub Actions.

### Currently not implemented

To run unit tests, run `make test`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:*
Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
