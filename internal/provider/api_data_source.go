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
	_ datasource.DataSource              = &apiDataSource{}
	_ datasource.DataSourceWithConfigure = &apiDataSource{}
)

// NewApiDataSource is a helper function to simplify the provider implementation.
func NewApiDataSource() datasource.DataSource {
	return &apiDataSource{}
}

// apiDataSource is the data source implementation.
type apiDataSource struct {
	client *WikiJSClient
}

type apiDataSourceModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// Metadata returns the data source type name.
func (d *apiDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api"
}

// Schema defines the schema for the data source.
func (d *apiDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "States whether the Wiki.JS API is enabled",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *apiDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *apiDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *apiDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetApiState(ctx, d.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get API State Request failed", err.Error())
		return
	}
	data.Enabled = types.BoolValue(wresp.Authentication.ApiState)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
