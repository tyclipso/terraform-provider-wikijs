# Use this data source to query the current setup and make changes in
# certain renderers. Note that this data source only takes an empty
# renderers list.
data "wikijs_theme_renderers" "renderers" {
	renderers = []
}
