# This data_source can be useful if you want to change only one of the
# required theme configs but don't want to change the rest

# Notice that this data source does not take any arguments
data "wikijs_theme_config" "config" {}

resource "wikijs_theme_config" "config" {
  theme   = data.wikijs_theme_config.config.theme
  iconset = data.wikijs_theme_config.data.iconset
  # Only switch dark mode globally on
  dark_mode = true
}
