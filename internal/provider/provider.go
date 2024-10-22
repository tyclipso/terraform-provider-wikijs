package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

// Ensure WikiJSProvider satisfies various provider interfaces.
var _ provider.Provider = &WikiJSProvider{}

// WikiJSProvider defines the provider implementation.
type WikiJSProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// WikiJSProviderModel describes the provider data model.
type WikiJSProviderModel struct {
	SiteUrl  types.String `tfsdk:"site_url"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
}

type WikiJSClient struct {
	http    *http.Client
	siteUrl *url.URL
	graphql graphql.Client
}

func (p *WikiJSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wikijs"
	resp.Version = p.version
}

func (p *WikiJSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"site_url": schema.StringAttribute{
				MarkdownDescription: "Wiki.JS Site URL",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email to login with",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password to login with",
				Optional:            true,
			},
		},
		MarkdownDescription: "The WikiJS provider aims to implement the complete GraphQL API of WikiJS.\n" +
			"It should be possible to configure any instance of WikiJS via this provider.\n" +
			"The main focus is for administration backend to be complete but also provide necessary resources and data_source to aid with that.\n" +
			"\n" +
			"Use the navigation to the left to read about the available resources.",
	}
}

func (p *WikiJSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var data WikiJSProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Email.IsNull() {
		data.Email = types.StringValue(os.Getenv("TF_PROVIDER_WIKIJS_EMAIL"))
	}

	if data.Password.IsNull() {
		data.Password = types.StringValue(os.Getenv("TF_PROVIDER_WIKIJS_PASSWORD"))
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		resp.Diagnostics.AddError("could not create cookie jar", err.Error())
		return
	}

	siteUrl, err := url.Parse(data.SiteUrl.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("could not parse site url", err.Error())
		return
	}

	client := &WikiJSClient{
		siteUrl: siteUrl,
		http: &http.Client{
			Jar: jar,
		},
	}
	client.graphql = graphql.NewClient(client.siteUrl.JoinPath("/graphql").String(), client.http)

	loginResp, err := wikijs.Login(ctx, client.graphql, data.Email.ValueString(), data.Password.ValueString(), "local")
	if err != nil {
		resp.Diagnostics.AddError("wiki.js login request failed", err.Error())
		return
	}

	if !loginResp.Authentication.Login.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("wiki.js Login failed: %s", loginResp.Authentication.Login.ResponseResult.Slug), loginResp.Authentication.Login.ResponseResult.Message)
		return
	}

	cookie := &http.Cookie{
		Name:  "jwt",
		Value: loginResp.Authentication.Login.Jwt,
	}
	client.http.Jar.SetCookies(client.siteUrl, []*http.Cookie{cookie})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *WikiJSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSiteConfigResource,
		NewGroupResource,
		NewManagedSystemGroupResource,
		NewLocalizationResource,
		NewPageResource,
		NewApiResource,
		NewApiKeyResource,
		NewAuthStrategiesResource,
		NewThemeConfigResource,
	}
}

func (p *WikiJSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSiteConfigDataSource,
		NewPageDataSource,
		NewGroupDataSource,
		NewGroupsDataSource,
		NewApiDataSource,
		NewApiKeysDataSource,
		NewAuthStrategiesDataSource,
		NewThemeConfigDataSource,
		NewThemesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WikiJSProvider{
			version: version,
		}
	}
}
