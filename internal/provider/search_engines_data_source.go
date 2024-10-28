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
	_ datasource.DataSource              = &searchEnginesDataSource{}
	_ datasource.DataSourceWithConfigure = &searchEnginesDataSource{}
)

func NewSearchEnginesDataSource() datasource.DataSource {
	return &searchEnginesDataSource{}
}

type searchEnginesDataSource struct {
	client *WikiJSClient
}

type searchEnginesDataModel struct {
	IsEnabled   types.Bool   `tfsdk:"is_enabled"`
	Key         types.String `tfsdk:"key"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Logo        types.String `tfsdk:"logo"`
	Website     types.String `tfsdk:"website"`
	IsAvailable types.Bool   `tfsdk:"is_available"`
	Config      types.Map    `tfsdk:"config"`
}

type searchEnginesDataSourceModel struct {
	SearchEngines []searchEnginesDataModel `tfsdk:"search_engines"`
}

func (d *searchEnginesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_search_engines"
}

func (d *searchEnginesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"search_engines": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of search_engines in the system.\n",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"is_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"key": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"logo": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"website": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"is_available": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "",
						},
						"config": schema.MapAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "",
						},
					},
				},
			},
		},
		MarkdownDescription: "",
	}
}

func (d *searchEnginesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *searchEnginesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state searchEnginesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetSearchEngines(ctx, d.client.graphql, "", "")
	if err != nil {
		resp.Diagnostics.AddError("Get Search Engines Request failed", err.Error())
	}

	type configValue struct {
		Value any `json:"value"`
	}

	state.SearchEngines = state.SearchEngines[:0]
	for i, s := range wresp.Search.SearchEngines {
		if len(state.SearchEngines) <= i {
			state.SearchEngines = append(state.SearchEngines, make([]searchEnginesDataModel, i+1-len(state.SearchEngines))...)
		}

		configTmp := map[string]string{}
		for _, c := range s.Config {
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

		state.SearchEngines[i].IsEnabled = types.BoolValue(s.IsEnabled)
		state.SearchEngines[i].Key = types.StringValue(s.Key)
		state.SearchEngines[i].Title = types.StringValue(s.Title)
		state.SearchEngines[i].Description = types.StringValue(s.Description)
		state.SearchEngines[i].Logo = types.StringValue(s.Logo)
		state.SearchEngines[i].Website = types.StringValue(s.Website)
		state.SearchEngines[i].IsAvailable = types.BoolValue(s.IsAvailable)
		state.SearchEngines[i].Config = config
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
