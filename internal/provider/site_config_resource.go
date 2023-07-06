package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.startnext.org/sre/terraform/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &siteConfigResource{}
	_ resource.ResourceWithConfigure = &siteConfigResource{}
)

// NewSiteConfigResource is a helper function to simplify the provider implementation.
func NewSiteConfigResource() resource.Resource {
	return &siteConfigResource{}
}

// siteConfigResource is the resource implementation.
type siteConfigResource struct {
	client *WikiJSClient
}

type siteConfigResourceModel struct {
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

// Metadata returns the resource type name.
func (r *siteConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_config"
}

// Schema defines the schema for the resource.
func (r *siteConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Full URL to your wiki, without the trailing slash. (e.g. https://wiki.example.com)",
				Default:     stringdefault.StaticString(""),
			},
			"title": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Displayed in the top bar and appended to all pages meta title.",
				Default:     stringdefault.StaticString("Wiki.js"),
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default description when none is provided for a page.",
				Default:     stringdefault.StaticString(""),
			},
			"robots": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default: Index, Follow. Can also be set on a per-page basis.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("index"), types.StringValue("follow")})),
			},
			"analytics_service": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"analytics_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"company": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name to use when displaying copyright notice in the footer. Leave empty to hide.",
				Default:     stringdefault.StaticString(""),
			},
			"content_license": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "License shown in the footer of all content pages.",
				Default:     stringdefault.StaticString(""),
			},
			"footer_override": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optionally override the footer text with a custom message. Useful if none of the above licenses are appropriate.",
				Default:     stringdefault.StaticString(""),
			},
			"logo_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Specify an image to use as the logo. SVG, PNG, JPG are supported, in a square ratio, 34x34 pixels or larger. Click the button on the right to upload a new image.",
				Default:     stringdefault.StaticString("https://static.requarks.io/logo/wikijs-butterfly.svg"),
			},
			"page_extensions": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A comma-separated list of URL extensions that will be treated as pages. For example, adding md will treat /foobar.md the same as /foobar.",
				Default:     stringdefault.StaticString("md, html, txt"),
			},
			"auth_auto_login": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Should the user be redirected automatically to the first authentication provider.",
				Default:     booldefault.StaticBool(false),
			},
			"auth_enforce_2fa": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Force all users to use Two-Factor Authentication when using an authentication provider with a user / password form.",
				Default:     booldefault.StaticBool(false),
			},
			"auth_hide_local": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Don't show the local authentication provider on the login screen. Add ?all to the URL to temporarily use it.",
				Default:     booldefault.StaticBool(false),
			},
			"auth_login_bg_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Specify an image to use as the login background. PNG and JPG are supported, 1920x1080 recommended. Leave empty for default. Click the button on the right to upload a new image. Note that the Guests group must have read-access to the selected image!",
				Default:     stringdefault.StaticString(""),
			},
			"auth_jwt_audience": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Audience URN used in JWT issued upon login. Usually your domain name. (e.g. urn:your.domain.com)",
				Default:     stringdefault.StaticString("urn:wiki.js"),
			},
			"auth_jwt_expiration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The expiration period of a token until it must be renewed. (default: 30m)",
				Default:     stringdefault.StaticString("30m"),
			},
			"auth_jwt_renewable_period": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The maximum period a token can be renewed when expired. (default: 14d)",
				Default:     stringdefault.StaticString("14d"),
			},
			"edit_fab": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display the edit floating action button (FAB) with a speed-dial menu in the bottom right corner of the screen.",
				Default:     booldefault.StaticBool(true),
			},
			"edit_menu_bar": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display the edit menu bar in the page header.",
				Default:     booldefault.StaticBool(false),
			},
			"edit_menu_btn": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"edit_menu_external_btn": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"edit_menu_external_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("GitHub"),
			},
			"edit_menu_external_icon": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("mdi-github"),
			},
			"edit_menu_external_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("https://github.com/org/repo/blob/main/{filename}"),
			},
			"feature_page_ratings": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"feature_page_comments": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Allow users to leave comments on pages.",
				Default:     booldefault.StaticBool(true),
			},
			"feature_personal_wikis": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"security_open_redirect": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Prevents user controlled URLs from directing to websites outside of your wiki. This provides Open Redirect protection.",
				Default:     booldefault.StaticBool(true),
			},
			"security_iframe": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Prevents other websites from embedding your wiki in an iframe. This provides clickjacking protection.",
				Default:     booldefault.StaticBool(true),
			},
			"security_referrer_policy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Limits the referrer header to same origin.",
				Default:     booldefault.StaticBool(true),
			},
			"security_trust_proxy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Should be enabled when using a reverse-proxy like nginx, apache, CloudFlare, etc in front of Wiki.js. Turn off otherwise.",
				Default:     booldefault.StaticBool(true),
			},
			"security_sri": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"security_hsts": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "This ensures the connection cannot be established through an insecure HTTP connection.",
				Default:     booldefault.StaticBool(false),
			},
			"security_hsts_duration": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines the duration for which the server should only deliver content through HTTPS. It's a good idea to start with small values and make sure that nothing breaks on your wiki before moving to longer values.",
				Default:     int64default.StaticInt64(300),
			},
			"security_csp": schema.BoolAttribute{
				Optional: true,
				Computed: true,

				Default: booldefault.StaticBool(false),
			},
			"security_csp_directives": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"upload_max_file_size": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The maximum size for a single file.",
				Default:     int64default.StaticInt64(5242880),
			},
			"upload_max_files": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "How many files can be uploaded in a single batch?",
				Default:     int64default.StaticInt64(10),
			},
			"upload_scan_svg": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Should SVG uploads be scanned for vulnerabilities and stripped of any potentially unsafe content.",
				Default:     booldefault.StaticBool(true),
			},
			"upload_force_download": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Should non-image files be forced as downloads when accessed directly. This prevents potential XSS attacks via unsafe file extensions uploads.",
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *siteConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*WikiJSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *WikiJSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *siteConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *siteConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var robots []string
	resp.Diagnostics.Append(data.Robots.ElementsAs(ctx, &robots, false)...)

	wresp, err := wikijs.UpdateSiteConfig(ctx, r.client.graphql, data.Host.ValueString(), data.Title.ValueString(), data.Description.ValueString(), robots, data.AnalyticsService.ValueString(), data.AnalyticsId.ValueString(), data.Company.ValueString(), data.ContentLicense.ValueString(), data.FooterOverride.ValueString(), data.LogoUrl.ValueString(), data.PageExtensions.ValueString(), data.AuthAutoLogin.ValueBool(), data.AuthEnforce2FA.ValueBool(), data.AuthHideLocal.ValueBool(), data.AuthLoginBgUrl.ValueString(), data.AuthJwtAudience.ValueString(), data.AuthJwtExpiration.ValueString(), data.AuthJwtRenewablePeriod.ValueString(), data.EditFab.ValueBool(), data.EditMenuBar.ValueBool(), data.EditMenuBtn.ValueBool(), data.EditMenuExternalBtn.ValueBool(), data.EditMenuExternalName.ValueString(), data.EditMenuExternalIcon.ValueString(), data.EditMenuExternalUrl.ValueString(), data.FeaturePageRatings.ValueBool(), data.FeaturePageComments.ValueBool(), data.FeaturePersonalWikis.ValueBool(), data.SecurityOpenRedirect.ValueBool(), data.SecurityIframe.ValueBool(), data.SecurityReferrerPolicy.ValueBool(), data.SecurityTrustProxy.ValueBool(), data.SecuritySRI.ValueBool(), data.SecurityHSTS.ValueBool(), int(data.SecurityHSTSDuration.ValueInt64()), data.SecurityCSP.ValueBool(), data.SecurityCSPDirectives.ValueString(), int(data.UploadMaxFileSize.ValueInt64()), int(data.UploadMaxFiles.ValueInt64()), data.UploadScanSVG.ValueBool(), data.UploadForceDownload.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("could not update wiki.js site config", err.Error())
		return
	}
	if !wresp.Site.UpdateConfig.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("wiki.js refused site config update: %s", wresp.Site.UpdateConfig.ResponseResult.Slug), wresp.Site.UpdateConfig.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *siteConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var state *siteConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetSiteConfig(ctx, r.client.graphql)
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

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *siteConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *siteConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var robots []string
	resp.Diagnostics.Append(data.Robots.ElementsAs(ctx, &robots, false)...)

	wresp, err := wikijs.UpdateSiteConfig(ctx, r.client.graphql, data.Host.ValueString(), data.Title.ValueString(), data.Description.ValueString(), robots, data.AnalyticsService.ValueString(), data.AnalyticsId.ValueString(), data.Company.ValueString(), data.ContentLicense.ValueString(), data.FooterOverride.ValueString(), data.LogoUrl.ValueString(), data.PageExtensions.ValueString(), data.AuthAutoLogin.ValueBool(), data.AuthEnforce2FA.ValueBool(), data.AuthHideLocal.ValueBool(), data.AuthLoginBgUrl.ValueString(), data.AuthJwtAudience.ValueString(), data.AuthJwtExpiration.ValueString(), data.AuthJwtRenewablePeriod.ValueString(), data.EditFab.ValueBool(), data.EditMenuBar.ValueBool(), data.EditMenuBtn.ValueBool(), data.EditMenuExternalBtn.ValueBool(), data.EditMenuExternalName.ValueString(), data.EditMenuExternalIcon.ValueString(), data.EditMenuExternalUrl.ValueString(), data.FeaturePageRatings.ValueBool(), data.FeaturePageComments.ValueBool(), data.FeaturePersonalWikis.ValueBool(), data.SecurityOpenRedirect.ValueBool(), data.SecurityIframe.ValueBool(), data.SecurityReferrerPolicy.ValueBool(), data.SecurityTrustProxy.ValueBool(), data.SecuritySRI.ValueBool(), data.SecurityHSTS.ValueBool(), int(data.SecurityHSTSDuration.ValueInt64()), data.SecurityCSP.ValueBool(), data.SecurityCSPDirectives.ValueString(), int(data.UploadMaxFileSize.ValueInt64()), int(data.UploadMaxFiles.ValueInt64()), data.UploadScanSVG.ValueBool(), data.UploadForceDownload.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("could not update wiki.js site config", err.Error())
		return
	}
	if !wresp.Site.UpdateConfig.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("wiki.js refused site config update: %s", wresp.Site.UpdateConfig.ResponseResult.Slug), wresp.Site.UpdateConfig.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *siteConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Site Config has no factory defaults", "Deleting the wikijs_site_config resource just removes the resource from the terraform state. No wiki.js config is changed.")
}
