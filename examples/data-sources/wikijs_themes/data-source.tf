# This data source currently only returns a list with one item. The
# themes query is hardcoded to only respond with the default theme in v2

data "wikijs_themes" "themes_list" {
  themes = []
}

# You could use it to reset the theme but it seems easier to just write
# "default"
resource "wikijs_theme_config" "theme_config" {
  theme     = data.wikijs_themes.themes_list[0].key
  iconset   = "mdi"
  dark_mode = false
}
