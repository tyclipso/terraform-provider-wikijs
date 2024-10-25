package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

var (
	_ datasource.DataSource              = &renderersDataSource{}
	_ datasource.DataSourceWithConfigure = &renderersDataSource{}
)

func NewRenderersDataSource() datasource.DataSource {
	return &renderersDataSource{}
}

type renderersDataSource struct {
	client *WikiJSClient
}

type renderersDataModel struct {
	IsEnabled   types.Bool   `tfsdk:"is_enabled"`
	Key         types.String `tfsdk:"key"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Icon        types.String `tfsdk:"icon"`
	DependsOn   types.String `tfsdk:"depends_on"`
	Input       types.String `tfsdk:"input"`
	Output      types.String `tfsdk:"output"`
	Config      types.Map    `tfsdk:"config"`
}

type renderersDataSourceModel struct {
	Renderers []renderersDataModel `tfsdk:"renderers"`
}

func (d *renderersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_renderers"
}

func (d *renderersDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"renderers": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of renderers in the system.\n",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"is_enabled": schema.BoolAttribute{
							Computed: true,
							MarkdownDescription: "Either if the renderer is active or not.\n" +
								"  You can use this field in the `renderers` ressource.\n",
						},
						"key": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "The unique identifier of each renderer.\n" +
								"  This is set in code and you can use this field in the `renderers` ressource.\n",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The title of the renderer shown in the backend.\n",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The description of the renderer shown in the backend.\n",
						},
						"icon": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The icon of the renderer shown in the backend.\n",
						},
						"depends_on": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "The `key` of the renderer, this renderer depends on.\n" +
								"  This could be `markdownCore` for a markdown renderer.\n" +
								"  The `renderers` ressource currently does not check against this, so make sure to activate them.\n",
						},
						"input": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "What kind of input format this renderer takes.\n" +
								"  This could be for example `markdown`, `asciidoc`, `openapi`, `html` or `null`.\n",
						},
						"output": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "What kind of output format this renderer produces.\n" +
								"  This could be for example `html` or `null`.\n",
						},
						"config": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
							MarkdownDescription: "A list of Key/Value pairs of config for each renderer.\n" +
								"  Some take none, others have a long list.\n" +
								"  You can use this field in the `renderers` ressource.\n",
						},
					},
				},
			},
		},
		MarkdownDescription: "The `{{ .Name }}` {{ .Type }} implements the WikiJS API query `rendering{renderers{…}}`.\n" +
			"You can use this data source to manipulate only certain fields with the `renderers` resource.\n",
	}
}

func (d *renderersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *renderersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state renderersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetRenderers(ctx, d.client.graphql, "", "")
	if err != nil {
		resp.Diagnostics.AddError("Get Renderers Request failed", err.Error())
	}

	type configValue struct {
		Value any `json:"value"`
	}

	state.Renderers = state.Renderers[:0]
	for i, r := range wresp.Rendering.Renderers {
		if len(state.Renderers) <= i {
			state.Renderers = append(state.Renderers, make([]renderersDataModel, i+1-len(state.Renderers))...)
		}

		configTmp := map[string]string{}
		for _, c := range r.Config {
			valueObj := configValue{}
			if err := json.Unmarshal([]byte(c.Value), &valueObj); err != nil {
				resp.Diagnostics.AddError("Could not unmarshal JSON", err.Error())
				return
			}
			switch value := valueObj.Value.(type) {
			case bool:
				configTmp[c.Key] = strconv.FormatBool(value)
			case json.Number:
				configTmp[c.Key] = value.String()
			case string:
				configTmp[c.Key] = value
			default:
				configTmp[c.Key] = fmt.Sprintf("%v", value)
			}
		}
		config, diag := types.MapValueFrom(ctx, types.StringType, configTmp)
		resp.Diagnostics.Append(diag...)

		state.Renderers[i].IsEnabled = types.BoolValue(r.IsEnabled)
		state.Renderers[i].Key = types.StringValue(r.Key)
		state.Renderers[i].Title = types.StringValue(r.Title)
		state.Renderers[i].Description = types.StringValue(r.Description)
		state.Renderers[i].Icon = types.StringValue(r.Icon)
		state.Renderers[i].DependsOn = types.StringValue(r.DependsOn)
		state.Renderers[i].Input = types.StringValue(r.Input)
		state.Renderers[i].Output = types.StringValue(r.Output)
		state.Renderers[i].Config = config
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
