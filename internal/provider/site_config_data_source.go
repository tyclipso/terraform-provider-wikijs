package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.startnext.org/sre/terraform/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &siteConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &siteConfigDataSource{}
)

// NewSiteConfigDataSource is a helper function to simplify the provider implementation.
func NewSiteConfigDataSource() datasource.DataSource {
	return &siteConfigDataSource{}
}

// siteConfigDataSource is the data source implementation.
type siteConfigDataSource struct {
	client *WikiJSClient
}

// siteConfigDataSourceModel maps the data source schema data.
type siteConfigDataSourceModel struct {
	Host                   types.String `tfsdk:"host"`
	Title                  types.String `tfsdk:"title"`
	Description            types.String `tfsdk:"description"`
	Robots                 types.List   `tfsdk:"robots"`
	AnalyticsService       types.String `tfsdk:"analytics_service"`
	AnalyticsId            types.String `tfsdk:"analytics_id"`
	Company                types.String `tfsdk:"company"`
	ContentLicense         types.String `tfsdk:"content_license"`
	FooterOverride         types.String `tfsdk:"footer_override"`
	LogoUrl                types.String `tfsdk:"logo_url"`
	PageExtensions         types.String `tfsdk:"page_extensions"`
	AuthAutoLogin          types.Bool   `tfsdk:"auth_auto_login"`
	AuthEnforce2FA         types.Bool   `tfsdk:"auth_enforce_2fa"`
	AuthHideLocal          types.Bool   `tfsdk:"auth_hide_local"`
	AuthLoginBgUrl         types.String `tfsdk:"auth_login_bg_url"`
	AuthJwtAudience        types.String `tfsdk:"auth_jwt_audience"`
	AuthJwtExpiration      types.String `tfsdk:"auth_jwt_expiration"`
	AuthJwtRenewablePeriod types.String `tfsdk:"auth_jwt_renewable_period"`
	EditFab                types.Bool   `tfsdk:"edit_fab"`
	EditMenuBar            types.Bool   `tfsdk:"edit_menu_bar"`
	EditMenuBtn            types.Bool   `tfsdk:"edit_menu_btn"`
	EditMenuExternalBtn    types.Bool   `tfsdk:"edit_menu_external_btn"`
	EditMenuExternalName   types.String `tfsdk:"edit_menu_external_name"`
	EditMenuExternalIcon   types.String `tfsdk:"edit_menu_external_icon"`
	EditMenuExternalUrl    types.String `tfsdk:"edit_menu_external_url"`
	FeaturePageRatings     types.Bool   `tfsdk:"feature_page_ratings"`
	FeaturePageComments    types.Bool   `tfsdk:"feature_page_comments"`
	FeaturePersonalWikis   types.Bool   `tfsdk:"feature_personal_wikis"`
	SecurityOpenRedirect   types.Bool   `tfsdk:"security_open_redirect"`
	SecurityIframe         types.Bool   `tfsdk:"security_iframe"`
	SecurityReferrerPolicy types.Bool   `tfsdk:"security_referrer_policy"`
	SecurityTrustProxy     types.Bool   `tfsdk:"security_trust_proxy"`
	SecuritySRI            types.Bool   `tfsdk:"security_sri"`
	SecurityHSTS           types.Bool   `tfsdk:"security_hsts"`
	SecurityHSTSDuration   types.Int64  `tfsdk:"security_hsts_duration"`
	SecurityCSP            types.Bool   `tfsdk:"security_csp"`
	SecurityCSPDirectives  types.String `tfsdk:"security_csp_directives"`
	UploadMaxFileSize      types.Int64  `tfsdk:"upload_max_file_size"`
	UploadMaxFiles         types.Int64  `tfsdk:"upload_max_files"`
	UploadScanSVG          types.Bool   `tfsdk:"upload_scan_svg"`
	UploadForceDownload    types.Bool   `tfsdk:"upload_force_download"`
}

// Metadata returns the data source type name.
func (d *siteConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_config"
}

// Schema defines the schema for the data source.
func (d *siteConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "Full URL to your wiki, without the trailing slash. (e.g. https://wiki.example.com)",
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: "Displayed in the top bar and appended to all pages meta title.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Default description when none is provided for a page.",
			},
			"robots": schema.ListAttribute{
				Computed:    true,
				Description: "Default: Index, Follow. Can also be set on a per-page basis.",
				ElementType: types.StringType,
			},
			"analytics_service": schema.StringAttribute{
				Computed: true,
			},
			"analytics_id": schema.StringAttribute{
				Computed: true,
			},
			"company": schema.StringAttribute{
				Computed:    true,
				Description: "Name to use when displaying copyright notice in the footer. Leave empty to hide.",
			},
			"content_license": schema.StringAttribute{
				Computed:    true,
				Description: "License shown in the footer of all content pages.",
			},
			"footer_override": schema.StringAttribute{
				Computed:    true,
				Description: "Optionally override the footer text with a custom message. Useful if none of the above licenses are appropriate.",
			},
			"logo_url": schema.StringAttribute{
				Computed:    true,
				Description: "Specify an image to use as the logo. SVG, PNG, JPG are supported, in a square ratio, 34x34 pixels or larger. Click the button on the right to upload a new image.",
			},
			"page_extensions": schema.StringAttribute{
				Computed:    true,
				Description: "A comma-separated list of URL extensions that will be treated as pages. For example, adding md will treat /foobar.md the same as /foobar.",
			},
			"auth_auto_login": schema.BoolAttribute{
				Computed:    true,
				Description: "Should the user be redirected automatically to the first authentication provider.",
			},
			"auth_enforce_2fa": schema.BoolAttribute{
				Computed:    true,
				Description: "Force all users to use Two-Factor Authentication when using an authentication provider with a user / password form.",
			},
			"auth_hide_local": schema.BoolAttribute{
				Computed:    true,
				Description: "Don't show the local authentication provider on the login screen. Add ?all to the URL to temporarily use it.",
			},
			"auth_login_bg_url": schema.StringAttribute{
				Computed:    true,
				Description: "Specify an image to use as the login background. PNG and JPG are supported, 1920x1080 recommended. Leave empty for default. Click the button on the right to upload a new image. Note that the Guests group must have read-access to the selected image!",
			},
			"auth_jwt_audience": schema.StringAttribute{
				Computed:    true,
				Description: "Audience URN used in JWT issued upon login. Usually your domain name. (e.g. urn:your.domain.com)",
			},
			"auth_jwt_expiration": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration period of a token until it must be renewed. (default: 30m)",
			},
			"auth_jwt_renewable_period": schema.StringAttribute{
				Computed:    true,
				Description: "The maximum period a token can be renewed when expired. (default: 14d)",
			},
			"edit_fab": schema.BoolAttribute{
				Computed:    true,
				Description: "Display the edit floating action button (FAB) with a speed-dial menu in the bottom right corner of the screen.",
			},
			"edit_menu_bar": schema.BoolAttribute{
				Computed:    true,
				Description: "Display the edit menu bar in the page header.",
			},
			"edit_menu_btn": schema.BoolAttribute{
				Computed: true,
			},
			"edit_menu_external_btn": schema.BoolAttribute{
				Computed: true,
			},
			"edit_menu_external_name": schema.StringAttribute{
				Computed: true,
			},
			"edit_menu_external_icon": schema.StringAttribute{
				Computed: true,
			},
			"edit_menu_external_url": schema.StringAttribute{
				Computed: true,
			},
			"feature_page_ratings": schema.BoolAttribute{
				Computed: true,
			},
			"feature_page_comments": schema.BoolAttribute{
				Computed:    true,
				Description: "Allow users to leave comments on pages.",
			},
			"feature_personal_wikis": schema.BoolAttribute{
				Computed: true,
			},
			"security_open_redirect": schema.BoolAttribute{
				Computed:    true,
				Description: "Prevents user controlled URLs from directing to websites outside of your wiki. This provides Open Redirect protection.",
			},
			"security_iframe": schema.BoolAttribute{
				Computed:    true,
				Description: "Prevents other websites from embedding your wiki in an iframe. This provides clickjacking protection.",
			},
			"security_referrer_policy": schema.BoolAttribute{
				Computed:    true,
				Description: "Limits the referrer header to same origin.",
			},
			"security_trust_proxy": schema.BoolAttribute{
				Computed:    true,
				Description: "Should be enabled when using a reverse-proxy like nginx, apache, CloudFlare, etc in front of Wiki.js. Turn off otherwise.",
			},
			"security_sri": schema.BoolAttribute{
				Computed: true,
			},
			"security_hsts": schema.BoolAttribute{
				Computed:    true,
				Description: "This ensures the connection cannot be established through an insecure HTTP connection.",
			},
			"security_hsts_duration": schema.Int64Attribute{
				Computed:    true,
				Description: "Defines the duration for which the server should only deliver content through HTTPS. It's a good idea to start with small values and make sure that nothing breaks on your wiki before moving to longer values.",
			},
			"security_csp": schema.BoolAttribute{
				Computed: true,
			},
			"security_csp_directives": schema.StringAttribute{
				Computed: true,
			},
			"upload_max_file_size": schema.Int64Attribute{
				Computed:    true,
				Description: "The maximum size for a single file.",
			},
			"upload_max_files": schema.Int64Attribute{
				Computed:    true,
				Description: "How many files can be uploaded in a single batch?",
			},
			"upload_scan_svg": schema.BoolAttribute{
				Computed:    true,
				Description: "Should SVG uploads be scanned for vulnerabilities and stripped of any potentially unsafe content.",
			},
			"upload_force_download": schema.BoolAttribute{
				Computed:    true,
				Description: "Should non-image files be forced as downloads when accessed directly. This prevents potential XSS attacks via unsafe file extensions uploads.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *siteConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*WikiJSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *WikiJSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *siteConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state siteConfigDataSourceModel

	wresp, err := wikijs.GetSiteConfig(ctx, d.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("could not query wiki.js graphql api", err.Error())
		return
	}

	state.Host = types.StringValue(wresp.Site.Config.Host)
	state.Title = types.StringValue(wresp.Site.Config.Title)
	state.Description = types.StringValue(wresp.Site.Config.Description)

	robots, diag := types.ListValueFrom(ctx, types.StringType, wresp.Site.Config.Robots)
	resp.Diagnostics = append(resp.Diagnostics, diag...)
	state.Robots = robots

	state.AnalyticsService = types.StringValue(wresp.Site.Config.AnalyticsService)
	state.AnalyticsId = types.StringValue(wresp.Site.Config.AnalyticsId)
	state.Company = types.StringValue(wresp.Site.Config.Company)
	state.ContentLicense = types.StringValue(wresp.Site.Config.ContentLicense)
	state.FooterOverride = types.StringValue(wresp.Site.Config.FooterOverride)
	state.LogoUrl = types.StringValue(wresp.Site.Config.LogoUrl)
	state.PageExtensions = types.StringValue(wresp.Site.Config.PageExtensions)
	state.AuthAutoLogin = types.BoolValue(wresp.Site.Config.AuthAutoLogin)
	state.AuthEnforce2FA = types.BoolValue(wresp.Site.Config.AuthEnforce2FA)
	state.AuthHideLocal = types.BoolValue(wresp.Site.Config.AuthHideLocal)
	state.AuthLoginBgUrl = types.StringValue(wresp.Site.Config.AuthLoginBgUrl)
	state.AuthJwtAudience = types.StringValue(wresp.Site.Config.AuthJwtAudience)
	state.AuthJwtExpiration = types.StringValue(wresp.Site.Config.AuthJwtExpiration)
	state.AuthJwtRenewablePeriod = types.StringValue(wresp.Site.Config.AuthJwtRenewablePeriod)
	state.EditFab = types.BoolValue(wresp.Site.Config.EditFab)
	state.EditMenuBar = types.BoolValue(wresp.Site.Config.EditMenuBar)
	state.EditMenuBtn = types.BoolValue(wresp.Site.Config.EditMenuBtn)
	state.EditMenuExternalBtn = types.BoolValue(wresp.Site.Config.EditMenuExternalBtn)
	state.EditMenuExternalName = types.StringValue(wresp.Site.Config.EditMenuExternalName)
	state.EditMenuExternalIcon = types.StringValue(wresp.Site.Config.EditMenuExternalIcon)
	state.EditMenuExternalUrl = types.StringValue(wresp.Site.Config.EditMenuExternalUrl)
	state.FeaturePageRatings = types.BoolValue(wresp.Site.Config.FeaturePageRatings)
	state.FeaturePageComments = types.BoolValue(wresp.Site.Config.FeaturePageComments)
	state.FeaturePersonalWikis = types.BoolValue(wresp.Site.Config.FeaturePersonalWikis)
	state.SecurityOpenRedirect = types.BoolValue(wresp.Site.Config.SecurityOpenRedirect)
	state.SecurityIframe = types.BoolValue(wresp.Site.Config.SecurityIframe)
	state.SecurityReferrerPolicy = types.BoolValue(wresp.Site.Config.SecurityReferrerPolicy)
	state.SecurityTrustProxy = types.BoolValue(wresp.Site.Config.SecurityTrustProxy)
	state.SecuritySRI = types.BoolValue(wresp.Site.Config.SecuritySRI)
	state.SecurityHSTS = types.BoolValue(wresp.Site.Config.SecurityHSTS)
	state.SecurityHSTSDuration = types.Int64Value(int64(wresp.Site.Config.SecurityHSTSDuration))
	state.SecurityCSP = types.BoolValue(wresp.Site.Config.SecurityCSP)
	state.SecurityCSPDirectives = types.StringValue(wresp.Site.Config.SecurityCSPDirectives)
	state.UploadMaxFileSize = types.Int64Value(int64(wresp.Site.Config.UploadMaxFileSize))
	state.UploadMaxFiles = types.Int64Value(int64(wresp.Site.Config.UploadMaxFiles))
	state.UploadScanSVG = types.BoolValue(wresp.Site.Config.UploadScanSVG)
	state.UploadForceDownload = types.BoolValue(wresp.Site.Config.UploadForceDownload)

	if !resp.Diagnostics.HasError() {
		// Set state
		diags := resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
	}
}
