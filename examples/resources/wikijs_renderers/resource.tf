# This is a _currently_ complete configuration for the renderers. If you
# add your own renderer, place the object in this list.
# The order of this example is as it can be found in the backend for
# easier working.

resource "wikijs_renderers" "wikijs_renderers" {
  renderers = [
    {
      key        = "openapiCore"
      is_enabled = true
      config     = {}
    },
    {
      key        = "markdownCore"
      is_enabled = true
      config = {
        "allowHTML"   = true
        "linkify"     = true
        "linebreaks"  = true
        "underline"   = false
        "typographer" = false
        # This field is an enum and has the following available values:
        # "Chinese", "English", "French", "German", "Greek", "Japanese",
        # "Hungarian", "Polish", "Portuguese", "Russian", "Spanish",
        # "Swedish"
        "quotes" = "English"
      }
    },
    {
      key        = "markdownAbbr"
      is_enabled = true
      config     = {}
    },
    {
      key        = "markdownEmoji"
      is_enabled = true
      config     = {}
    },
    {
      key        = "markdownExpandtabs"
      is_enabled = true
      config = {
        "tabWidth" = 4
      }
    },
    {
      key        = "markdownFootnotes"
      is_enabled = true
      config     = {}
    },
    {
      key        = "markdownImsize"
      is_enabled = true
      config     = {}
    },
    {
      key        = "markdownKatex"
      is_enabled = true
      config = {
        "useInline" : true
        "useBlocks" : true
      }
    },
    {
      key        = "markdownKroki"
      is_enabled = false
      config = {
        "server"      = "https://kroki.io"
        "openMarker"  = "```kroki"
        "closeMarker" = "```"
      }
    },
    {
      key        = "markdownMathjax"
      is_enabled = false
      config = {
        "useInline" = true
        "useBlocks" = true
      }
    },
    {
      key        = "markdownMultiTable"
      is_enabled = false
      config = {
        "headerlessEnabled" = true
        "multilineEnabled"  = true
        "rowspanEnabled"    = true
      }
    },
    {
      key        = "markdownPivotTable"
      is_enabled = false
      config     = {}
    },
    {
      key        = "markdownPlantuml"
      is_enabled = true
      config = {
        "server"      = "https://plantuml.requarks.io"
        "openMarker"  = "```plantuml"
        "closeMarker" = "```"
        "imageFormat" = "svg"
      }
    },
    {
      key        = "markdownSupsub"
      is_enabled = true
      config = {
        "subEnabled" = true
        "supEnabled" = true
      }
    },
    {
      key        = "markdownTasklists"
      is_enabled = true
      config     = {}
    },
    {
      key        = "asciidocCore"
      is_enabled = true
      config = {
        # This field is an enum and has the following available values:
        # "unsafe", "safe", "server", "secure"
        "safeMode" = "server"
      }
    },
    {
      key        = "htmlCore"
      is_enabled = true
      config = {
        "absoluteLinks"          = false
        "openExternalLinkNewTab" = false
        # This field is an enum and has the following available values:
        # "noreferrer", "noopener"
        "relAttributeExternalLink" = "noreferrer"
      }
    },
    {
      key        = "htmlAsciinema"
      is_enabled = false
      config     = {}
    },
    {
      key        = "htmlBlockquotes"
      is_enabled = true
      config     = {}
    },
    {
      key        = "htmlCodehighlighter"
      is_enabled = true
      config     = {}
    },
    {
      key        = "htmlDiagram"
      is_enabled = true
      config     = {}
    },
    {
      key        = "htmlImagePrefetch"
      is_enabled = false
      config     = {}
    },
    {
      key        = "htmlMediaplayers"
      is_enabled = true
      config     = {}
    },
    {
      key        = "htmlMermaid"
      is_enabled = true
      config     = {}
    },
    {
      key        = "htmlSecurity"
      is_enabled = true
      config = {
        "safeHTML"          = true
        "allowDrawIoUnsafe" = true
        "allowIFrames"      = false
      }
    },
    {
      key        = "htmlTabset"
      is_enabled = true
      config     = {}
    },
    {
      key        = "htmlTwemoji"
      is_enabled = true
      config     = {}
    }
  ]
}
