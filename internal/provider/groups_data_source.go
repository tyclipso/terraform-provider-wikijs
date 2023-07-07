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
	_ datasource.DataSource              = &groupsDataSource{}
	_ datasource.DataSourceWithConfigure = &groupsDataSource{}
)

// NewGroupsDataSource is a helper function to simplify the provider implementation.
func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

// groupsDataSource is the data source implementation.
type groupsDataSource struct {
	client *WikiJSClient
}

// groupsDataSourceModel maps the data source schema data.
type groupsDataSourceModel struct {
	Filter  types.String        `tfsdk:"filter"`
	OrderBy types.String        `tfsdk:"order_by"`
	Groups  []groupMinimalModel `tfsdk:"groups"`
}

type groupMinimalModel struct {
	Id        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	IsSystem  types.Bool   `tfsdk:"is_system"`
	UserCount types.Int64  `tfsdk:"user_count"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *groupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

// Schema defines the schema for the data source.
func (d *groupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filter": schema.StringAttribute{
				Optional:    true,
				Description: "Seems like this is just part of the graphql schema but not implemented in the wiki.js server",
			},
			"order_by": schema.StringAttribute{
				Optional:    true,
				Description: "Seems like this is just part of the graphql schema but not implemented in the wiki.js server",
			},
			"groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Internal id of the group.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the group",
						},
						"is_system": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is a system group",
						},
						"user_count": schema.Int64Attribute{
							Computed:    true,
							Description: "Number of users in this group",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation time of this group (expect RFC3399)",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "Last update time of this group (expect RFC3399)",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *groupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetGroups(ctx, d.client.graphql, state.Filter.ValueString(), state.OrderBy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Get Group List Query failed", err.Error())
		return
	}

	for _, g := range wresp.Groups.List {
		state.Groups = append(state.Groups, groupMinimalModel{
			Id:        types.Int64Value(int64(g.Id)),
			Name:      types.StringValue(g.Name),
			IsSystem:  types.BoolValue(g.IsSystem),
			UserCount: types.Int64Value(int64(g.UserCount)),
			CreatedAt: types.StringValue(g.CreatedAt),
			UpdatedAt: types.StringValue(g.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
