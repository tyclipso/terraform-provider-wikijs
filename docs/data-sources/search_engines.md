---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "wikijs_search_engines Data Source - terraform-provider-wikijs"
subcategory: ""
description: |-
  The wikijs_search_engines Data Source implements the WikiJS API query search{searchEngines{…}}.
  You can use this Data Source to manipulate only certain fields with the search_engines resource.
---

# wikijs_search_engines (Data Source)

The `wikijs_search_engines` Data Source implements the WikiJS API query `search{searchEngines{…}}`.
You can use this Data Source to manipulate only certain fields with the `search_engines` resource.

## Example Usage

```terraform
# Use this data source to query the current setup and make changes in
# certain search engindes. Note that this data source only takes an
# empty search_engines list.
data "wikijs_search_engines" "search_engines" {
  search_engines = []
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `search_engines` (Attributes List) List of search engines in the system. (see [below for nested schema](#nestedatt--search_engines))

<a id="nestedatt--search_engines"></a>
### Nested Schema for `search_engines`

Read-Only:

- `config` (Map of String, Sensitive) A list of Key/Value pairs of config for each search engine.
  Some take none, others have a long list.
  You can use this field in the `search_engines` resource.
- `description` (String) The description of the search engines shown in the backend.
- `is_available` (Boolean) Wether the implementation of this search engine is finished and can be used.
  Check this field before enabling a search engine with the resource `wikijs_search_engines`.
- `is_enabled` (Boolean) Either if the search engine is active or not.
  You can use this field in the `search_engines` resource.
- `key` (String) The unique identifier of each search engine.
  This is set in code and you can use this field in the `search_engines` resource.
- `logo` (String) The logo of the search engine shown in the backend.
- `title` (String) The title of the search engine shown in the backend.
- `website` (String) The website of the search engine.


