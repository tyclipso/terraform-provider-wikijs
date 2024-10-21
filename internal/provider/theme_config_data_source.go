package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

var (
	_ datasource.DataSource              = &themeConfigDataSource{}
	_ datasource.DataSourceWithConfigure = &themeConfigDataSource{}
)

func NewThemeConfigDataSource() datasource.DataSource {
	return &themeConfigDataSource{}
}

type themeConfigDataSource struct {
	client *WikiJSClient
}

type themeConfigDataSourceModel struct {
	Theme       types.String `tfsdk:"theme"`
	Iconset     types.String `tfsdk:"iconset"`
	DarkMode    types.Bool   `tfsdk:"dark_mode"`
	TocPosition types.String `tfsdk:"toc_position"`
	InjectCSS   types.String `tfsdk:"inject_css"`
	InjectHead  types.String `tfsdk:"inject_head"`
	InjectBody  types.String `tfsdk:"inject_body"`
}

func (d *themeConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_theme_config"
}

func (d *themeConfigDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"theme": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "Themes affect how content pages are displayed.\n" +
					"  Other site sections (such as the editor or admin area) are not affected.",
			},
			"iconset": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Set of icons to use for the sidebar navigation.",
			},
			"dark_mode": schema.BoolAttribute{
				Computed: true,
				MarkdownDescription: "Dark Mode.\n" +
					"  Not recommended for accessibility.\n" +
					"  May not be supported by all themes.",
			},
			"toc_position": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Select whether the table of contents is shown on the left, right or not at all.",
			},
			"inject_css": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "CSS code to inject after system default CSS.\n" +
					"  Consider using custom themes if you have a large amount of css code.\n" +
					"  Injecting too much CSS code will result in poor page load performance!\n" +
					"  CSS will automatically be minified.\n" +
					"  \n" +
					"  **CAUTION**: When adding styles for page content, you must scope them to the `.contents` class.\n" +
					"  Omitting this could break the layout of the editor!",
			},
			"inject_head": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "HTML code to be injected just before the closing head tag." +
					"  Usually for script tags.",
			},
			"inject_body": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "HTML code to be injected just before the closing body tag.",
			},
		},
		MarkdownDescription: "The `{{ .Name }}` {{ .Type }} implements the WikiJS API query `theming{config{…}}`.\n" +
			"It can be used to read the current state and only change one of the required or any of the optional, without touching the required.\n" +
			"The Schema descriptions are directly lifted from the descriptions of the input fields in WikiJS.",
	}
}

func (d *themeConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *themeConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state themeConfigDataSourceModel

	wresp, err := wikijs.GetThemeConfig(ctx, d.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get Theme Config Request failed", err.Error())
	}

	state.Theme = types.StringValue(wresp.Theming.Config.Theme)
	state.Iconset = types.StringValue(wresp.Theming.Config.Iconset)
	state.DarkMode = types.BoolValue(wresp.Theming.Config.DarkMode)
	state.TocPosition = types.StringValue(wresp.Theming.Config.TocPosition)
	state.InjectCSS = types.StringValue(wresp.Theming.Config.InjectCSS)
	state.InjectHead = types.StringValue(wresp.Theming.Config.InjectHead)
	state.InjectBody = types.StringValue(wresp.Theming.Config.InjectBody)

	if !resp.Diagnostics.HasError() {
		diags := resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
	}
}
