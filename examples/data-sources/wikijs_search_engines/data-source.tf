# Use this data source to query the current setup and make changes in
# certain search engindes. Note that this data source only takes an
# empty search_engines list.
data "wikijs_search_engines" "search_engines" {
  search_engines = []
}
