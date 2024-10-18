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
	_ datasource.DataSource              = &themesDataSource{}
	_ datasource.DataSourceWithConfigure = &themesDataSource{}
)

func NewThemesDataSource() datasource.DataSource {
	return &themesDataSource{}
}

type themesDataSource struct {
	client *WikiJSClient
}

type themesModel struct {
	Key    types.String `tfsdk:"key"`
	Title  types.String `tfsdk:"title"`
	Author types.String `tfsdk:"author"`
}

type themesDataSourceModel struct {
	Themes []themesModel `tfsdk:"themes"`
}

func (d *themesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_themes"
}

func (d *themesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"themes": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of registered themes. In wikijs v2 this contains only the default theme",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed: true,
						},
						"title": schema.StringAttribute{
							Computed: true,
						},
						"author": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *themesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *themesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state themesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wresp, err := wikijs.GetThemes(ctx, d.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get Themes Request failed", err.Error())
	}

	// Clear state.Themes to reassign values
	state.Themes = state.Themes[:0]
	for i, t := range wresp.Theming.Themes {
		if len(state.Themes) <= i {
			state.Themes = append(state.Themes, make([]themesModel, i+1-len(state.Themes))...)
		}
		state.Themes[i].Key = types.StringValue(t.Key)
		state.Themes[i].Title = types.StringValue(t.Title)
		state.Themes[i].Author = types.StringValue(t.Author)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
