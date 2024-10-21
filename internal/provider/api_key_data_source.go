package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &apiKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &apiKeysDataSource{}
)

// NewApiKeysDataSource is a helper function to simplify the provider implementation.
func NewApiKeysDataSource() datasource.DataSource {
	return &apiKeysDataSource{}
}

// apiKeysDataSource is the data source implementation.
type apiKeysDataSource struct {
	client *WikiJSClient
}

// apiKeysDataSourceModel maps the data source schema data.
type apiKeysDataSourceModel struct {
	ApiKey []apiKeyModel `tfsdk:"api_keys"`
}

type apiKeyModel struct {
	Id         types.Int64  `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	KeyShort   types.String `tfsdk:"key_short"`
	Expiration types.String `tfsdk:"expiration"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
	IsRevoked  types.Bool   `tfsdk:"is_revoked"`
}

// Metadata returns the data source type name.
func (d *apiKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_keys"
}

// Schema defines the schema for the data source.
func (d *apiKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_keys": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of all API Keys known to wiki.js, see the nested object for details.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Internal Id of the API Key",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name of the API Key (not unique)",
						},
						"key_short": schema.StringAttribute{
							Computed:    true,
							Description: "Suffix of the actual API Key as shown in the Web UI (Hint: Wiki.JS always adds three dots before the actual suffix. You need to remove those if you need to match an actual key)",
						},
						"expiration": schema.StringAttribute{
							Computed:    true,
							Description: "Expiration time as returned by Wiki.JS (expect RFC3339 format)",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation time as returned by Wiki.JS (expect RFC3339 format)",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "Time of last updated of the key as returned by Wiki.JS (expect RFC3339 format)",
						},
						"is_revoked": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the API Key is revoked",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *apiKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *apiKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apiKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetApiKeys(ctx, d.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get API Keys Request failed", err.Error())
		return
	}
	for _, k := range wresp.Authentication.ApiKeys {
		state.ApiKey = append(state.ApiKey, apiKeyModel{
			Id:         types.Int64Value(int64(k.Id)),
			Name:       types.StringValue(k.Name),
			KeyShort:   types.StringValue(k.KeyShort),
			Expiration: types.StringValue(k.Expiration),
			CreatedAt:  types.StringValue(k.CreatedAt),
			UpdatedAt:  types.StringValue(k.UpdatedAt),
			IsRevoked:  types.BoolValue(k.IsRevoked),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
