package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

var (
	defaultThemeConfig = themeConfig{
		theme:       "default",
		iconset:     "mdi",
		darkMode:    false,
		tocPosition: "left",
		injectCSS:   "",
		injectHead:  "",
		injectBody:  "",
	}
	_ resource.Resource                   = &themeConfigResource{}
	_ resource.ResourceWithConfigure      = &themeConfigResource{}
	_ resource.ResourceWithValidateConfig = &themeConfigResource{}
)

func NewThemeConfigResource() resource.Resource {
	return &themeConfigResource{}
}

type themeConfigResource struct {
	client *WikiJSClient
}

type themeConfigResourceModel struct {
	Theme       types.String `tfsdk:"theme"`
	Iconset     types.String `tfsdk:"iconset"`
	DarkMode    types.Bool   `tfsdk:"dark_mode"`
	TocPosition types.String `tfsdk:"toc_position"`
	InjectCSS   types.String `tfsdk:"inject_css"`
	InjectHead  types.String `tfsdk:"inject_head"`
	InjectBody  types.String `tfsdk:"inject_body"`
}

type themeConfig struct {
	theme       string
	iconset     string
	darkMode    bool
	tocPosition string
	injectCSS   string
	injectHead  string
	injectBody  string
}

func (r *themeConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme_config"
}

func (r *themeConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"theme": schema.StringAttribute{
				Required:    true,
				Description: "Themes affect how content pages are displayed. Other site sections (such as the editor or admin area) are not affected.",
			},
			"iconset": schema.StringAttribute{
				Required:    true,
				Description: "Set of icons to use for the sidebar navigation. Values: mdi, fa, fa4",
			},
			"dark_mode": schema.BoolAttribute{
				Required:    true,
				Description: "Dark Mode. Not recommended for accessibility. May not be supported by all themes.",
			},
			"toc_position": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Select whether the table of contents is shown on the \"left\", \"right\" or not at all (\"off\").",
				Default:     stringdefault.StaticString(defaultThemeConfig.tocPosition),
			},
			"inject_css": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "CSS code to inject after system default CSS. Consider using custom themes if you have a large amount of css code.\nInjecting too much CSS code will result in poor page load performance! CSS will automatically be minified.\nCAUTION: When adding styles for page content, you must scope them to the .contents class. Omitting this could break the layout of the editor!",
				Default:     stringdefault.StaticString(defaultThemeConfig.injectCSS),
			},
			"inject_head": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "HTML code to be injected just before the closing head tag. Usually for script tags.",
				Default:     stringdefault.StaticString(defaultThemeConfig.injectHead),
			},
			"inject_body": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "HTML code to be injected just before the closing body tag.",
				Default:     stringdefault.StaticString(defaultThemeConfig.injectBody),
			},
		},
	}
}

func (r *themeConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*WikiJSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expexted *WikiJSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *themeConfigResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data *themeConfigResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	iconset := []string{"mdi", "fa", "fa4"}
	if !slices.Contains(iconset, data.Iconset.ValueString()) {
		s := strings.Join(iconset, ", ")
		resp.Diagnostics.AddAttributeError(
			path.Root("iconset"),
			"Attribute Configured Wrong",
			fmt.Sprintf("Expected iconset to be one of %s", s),
		)
		return
	}

	tocPosition := []string{"left", "right", "off"}
	if !slices.Contains(tocPosition, data.TocPosition.ValueString()) {
		s := strings.Join(tocPosition, ", ")
		resp.Diagnostics.AddAttributeError(
			path.Root("toc_position"),
			"Attribute Configured Wrong",
			fmt.Sprintf("Expected toc_position to be one of %s", s),
		)
		return
	}

}

func (r *themeConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *themeConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetThemeConfig(ctx, r.client.graphql, data.Theme.ValueString(), data.Iconset.ValueString(), data.DarkMode.ValueBool(), data.TocPosition.ValueString(), data.InjectCSS.ValueString(), data.InjectHead.ValueString(), data.InjectBody.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not update wiki.js theme config", err.Error())
		return
	}
	if !wresp.Theming.SetConfig.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("wiki.js refused theme config update: %s", wresp.Theming.SetConfig.ResponseResult.Slug), wresp.Theming.SetConfig.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *themeConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *themeConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetThemeConfig(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Could not query wiki.js graphql api", err.Error())
	}
	state.Theme = types.StringValue(wresp.Theming.Config.Theme)
	state.Iconset = types.StringValue(wresp.Theming.Config.Iconset)
	state.DarkMode = types.BoolValue(wresp.Theming.Config.DarkMode)
	state.TocPosition = types.StringValue(wresp.Theming.Config.TocPosition)
	state.InjectCSS = types.StringValue(wresp.Theming.Config.InjectCSS)
	state.InjectHead = types.StringValue(wresp.Theming.Config.InjectHead)
	state.InjectBody = types.StringValue(wresp.Theming.Config.InjectBody)

	resp.Diagnostics.Append(req.State.Set(ctx, state)...)
}

func (r *themeConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *themeConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetThemeConfig(ctx, r.client.graphql, data.Theme.ValueString(), data.Iconset.ValueString(), data.DarkMode.ValueBool(), data.TocPosition.ValueString(), data.InjectCSS.ValueString(), data.InjectHead.ValueString(), data.InjectBody.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not update wiki.js theme config", err.Error())
		return
	}
	if !wresp.Theming.SetConfig.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("wiki.js refused theme config update: %s", wresp.Theming.SetConfig.ResponseResult.Slug), wresp.Theming.SetConfig.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *themeConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *themeConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetThemeConfig(ctx, r.client.graphql, defaultThemeConfig.theme, defaultThemeConfig.iconset, defaultThemeConfig.darkMode, defaultThemeConfig.tocPosition, defaultThemeConfig.injectCSS, defaultThemeConfig.injectHead, defaultThemeConfig.injectBody)
	if err != nil {
		resp.Diagnostics.AddError("Request to reset the Theme Config to Defaults has failed", err.Error())
		return
	}
	if !wresp.Theming.SetConfig.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not reset the Theme Config to Defaults: %s", wresp.Theming.SetConfig.ResponseResult.Slug), wresp.Theming.SetConfig.ResponseResult.Message)
		return
	}
	//resp.Diagnostics.AddWarning("Theme Config has no factory defaults", "Deleting the wikijs_theme_config resource just removes the resource from the terraform state. No wiki.js config is changed.")
}
