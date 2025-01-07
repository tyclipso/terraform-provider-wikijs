# Create an API key that is valid for 30 days with full access/admin
# access

resource "wikijs_api_key" "admin_key" {
  expires_in  = "30d"
  name        = "Wiki.JS Adminkey"
  full_access = true
}

# Create an API key that only allows group 3 level access that is one
# year old

resource "wikijs_api_key" "group_key" {
  expires_in = "1y"
  name       = "Wiki.JS Grouplevelkey"
  group_id   = "3"
}

# Create an API key that gets renewed/recreated through terraform a
# month prior to expiration

resource "wikijs_api_key" "renewed_key" {
	expires_in = "1y"
	name = "Wiki.JS Renewing Key"
	full_access = true
	min_remaining_duration = "1m"
}
