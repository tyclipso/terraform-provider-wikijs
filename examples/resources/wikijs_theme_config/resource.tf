# The following example shows how to configure the wikijs theme. Since
# v2 of Wiki.js does not allow for dynamic addition of custom themes you
# need to provide your own string there

resource "wikijs_theme_config" "config" {
  theme     = "custom_name"
  iconset   = "mdi"
  dark_mode = true
}

# If you want to change the iconset currently the provider only supports
# the builtin names:
# Material Design Icons: "mdi"
# Font Awesome: "fa"
# Font Awesome 4: "fa4"

resource "wikijs_theme_config" "config" {
  theme     = "default"
  iconset   = "mdi"
  dark_mode = false
}

# The provider has a builtin default theme_config that is set with the
# WikiJS setup. This config will be used for any Delete and a resource
# representation looks like this

resource "wikijs_theme_config" "config" {
  theme        = "default"
  iconset      = "mdi"
  dark_mode    = false
  toc_position = "left"
  inject_css   = ""
  inject_head  = ""
  inject_body  = ""
}
