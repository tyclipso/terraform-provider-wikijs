# This is a _currently_ complete configuration for the search_engines.
# The order of this example is as it can be found in the backend for
# easier working.

resource "wikijs_search_engines" "search_engines" {
  search_engines = [
    {
      key        = "aws"
      is_enabled = false
      config = {
        "domain"   = ""
        "endpoint" = ""
        # This field is an enum and has the following available values:
        # "ap-northeast-1", "ap-northeast-2", "ap-southeast-1",
        # "ap-southeast-2", "eu-central-1", "eu-west-1", "sa-east-1",
        # "us-east-1", "us-west-1", "us-west-2"
        "region"          = "us-east-1"
        "accessKeyId"     = ""
        "secretAccessKey" = ""
        # This field is an enum and has the following available values:
        # "ar", "bg", "ca", "cs", "da", "de", "el", "en", "es", "eu",
        # "fa", "fi", "fr", "ga", "gl", "he", "hi", "hu", "hy", "id",
        # "it", "ja", "ko", "lv", "mul", "nl", "no", "pt", "ro", "ru",
        # "sw", "th", "tr", "zh-Hans", "zh-Hant"
        "AnalysisSchemeLang" = "en"
      }
    },
    {
      key        = "algolia"
      is_enabled = false
      config = {
        "appId"     = ""
        "key"       = ""
        "indexName" = "wiki"
      }
    },
    {
      key        = "azure"
      is_enabled = false
      config = {
        "serviceName" = ""
        "adminKey"    = ""
        "indexName"   = "wiki"
      }
    },
    {
      key        = "db"
      is_enabled = true
      config     = {}
    },
    {
      key        = "postgres"
      is_enabled = false
      config = {
        # This field is an enum and has the following available values:
        # "simple", "danish", "dutch", "english", "finnish", "french",
        # "german", "hungarian", "italian", "norwegian", "portugese",
        # "romanian", "russian", "spanish", "swedish", "turkish"
        "dictLanguage" = "english"
      }
    },
    {
      key        = "elasticsearch"
      is_enabled = false
      config = {
        # This field is an enum and has the following available values:
        # "7.x", "6.x"
        "apiVersion"           = "6.x"
        "hosts"                = ""
        "verifyTLSCertificate" = true
        "tlsCertPath"          = ""
        "indexName"            = "wiki"
        "analyzer"             = "simple"
        "sniffOnStart"         = false
        "sniffInterval"        = 0
      }
    },
    {
      key = "manticore"
      # This search engine is currently not available and therefore
      # cannot be enabled
      is_enabled = false
      config     = {}
    },
    {
      key = "solr"
      # This search engine is currently not available and therefore
      # cannot be enabled
      is_enabled = false
      config = {
        "host"     = "solr"
        "port"     = "8983"
        "core"     = "wiki"
        "protocol" = "http"
      }
    },
    {
      key = "sphinx"
      # This search engine is currently not available and therefore
      # cannot be enabled
      is_enabled = false
      config     = {}
    }
  ]
}
