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
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

// NewGroupDataSource is a helper function to simplify the provider implementation.
func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

// groupDataSource is the data source implementation.
type groupDataSource struct {
	client *WikiJSClient
}

// groupDataSourceModel maps the data source schema data.
type groupDataSourceModel struct {
	Id              types.Int64          `tfsdk:"group_id"`
	Name            types.String         `tfsdk:"name"`
	IsSystem        types.Bool           `tfsdk:"is_system"`
	RedirectOnLogin types.String         `tfsdk:"redirect_on_login"`
	Permissions     types.List           `tfsdk:"permissions"`
	PageRules       []groupPageRuleModel `tfsdk:"page_rules"`
	CreatedAt       types.String         `tfsdk:"created_at"`
	UpdatedAt       types.String         `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the data source.
func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group_id": schema.Int64Attribute{
				Required:    true,
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
			"redirect_on_login": schema.StringAttribute{
				Computed:    true,
				Description: "Path to redirect members to upon login",
			},
			"permissions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Global permissions for this group (see: https://github.com/requarks/wiki/blob/db8a09fe8c267a54fbbfabe0dc871a2108824968/client/components/admin/admin-groups-edit-permissions.vue#L43)",
			},
			"page_rules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Page rules for this group. See nested object",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Internal ID for this rule",
						},
						"deny": schema.BoolAttribute{
							Computed:    true,
							Description: "Defines whether this is a deny or allow rule",
						},
						"match": schema.StringAttribute{
							Computed:            true,
							Description:         "Match pattern for this rule.",
							MarkdownDescription: "Match pattern for this rule. One of:\n  - START\n  - EXACT\n  - END\n  - REGEX\n  - TAG",
						},
						"roles": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Permissions of this role (see: https://github.com/requarks/wiki/blob/db8a09fe8c267a54fbbfabe0dc871a2108824968/client/components/admin/admin-groups-edit-permissions.vue#L43)",
						},
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "Path to match on",
						},
						"locales": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Locale to match on. Empty list to match on any locale.",
						},
					},
				},
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetGroup(ctx, d.client.graphql, int(state.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Get Group Query failed", err.Error())
		return
	}

	state.Name = types.StringValue(wresp.Groups.Single.Name)
	state.IsSystem = types.BoolValue(wresp.Groups.Single.IsSystem)
	state.RedirectOnLogin = types.StringValue(wresp.Groups.Single.RedirectOnLogin)

	permissions, diag := types.ListValueFrom(ctx, types.StringType, wresp.Groups.Single.Permissions)
	resp.Diagnostics = append(resp.Diagnostics, diag...)
	state.Permissions = permissions

	state.PageRules = []groupPageRuleModel{}
	for _, r := range wresp.Groups.Single.PageRules {
		roles, diag := types.ListValueFrom(ctx, types.StringType, r.Roles)
		resp.Diagnostics = append(resp.Diagnostics, diag...)
		locales, diag := types.ListValueFrom(ctx, types.StringType, r.Locales)
		resp.Diagnostics = append(resp.Diagnostics, diag...)

		state.PageRules = append(state.PageRules, groupPageRuleModel{
			Id:      types.StringValue(r.Id),
			Deny:    types.BoolValue(r.Deny),
			Match:   types.StringValue(string(r.Match)),
			Roles:   roles,
			Path:    types.StringValue(r.Path),
			Locales: locales,
		})
	}

	state.CreatedAt = types.StringValue(wresp.Groups.Single.CreatedAt)
	state.UpdatedAt = types.StringValue(wresp.Groups.Single.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
