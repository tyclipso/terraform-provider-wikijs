# Minimal example how to use the provider

terraform {
  required_providers {
    wikijs = {
      source  = "tyclipso/wikijs"
      version = "~> 1"
    }
  }
}

provider "wikijs" {
  site_url = "https://wiki.example.com"
}
