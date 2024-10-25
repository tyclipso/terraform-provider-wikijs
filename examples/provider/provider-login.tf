# If you want to edit content or administer the wiki
# you need to login

terraform {
  required_providers {
    wikijs = {
      source  = "tyclipso/wikijs"
      version = "~> 1"
    }
    random = {
      source = "hashicorp/random"
    }
  }
}

# You can generate a random password and send it to the /finalize
# endpoint after starting a wikijs instance
resource "random_password" "wikijs_admin_password" {
  length = 32
}

provider "wikijs" {
  site_url = "https://wiki.example.com"
  email    = "admin@wiki.example.com"
  password = random_password.wikijs_admin_password.result
}
