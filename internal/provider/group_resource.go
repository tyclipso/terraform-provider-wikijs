package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &groupResource{}
	_ resource.ResourceWithConfigure = &groupResource{}
)

// NewGroupResource is a helper function to simplify the provider implementation.
func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// groupResource is the resource implementation.
type groupResource struct {
	client *WikiJSClient
}

type groupResourceModel struct {
	Id              types.Int64          `tfsdk:"id"`
	Name            types.String         `tfsdk:"name"`
	IsSystem        types.Bool           `tfsdk:"is_system"`
	RedirectOnLogin types.String         `tfsdk:"redirect_on_login"`
	Permissions     types.List           `tfsdk:"permissions"`
	PageRules       []groupPageRuleModel `tfsdk:"page_rules"`
}

type groupPageRuleModel struct {
	Id      types.String `tfsdk:"id"`
	Deny    types.Bool   `tfsdk:"deny"`
	Match   types.String `tfsdk:"match"`
	Roles   types.List   `tfsdk:"roles"`
	Path    types.String `tfsdk:"path"`
	Locales types.List   `tfsdk:"locales"`
}

// Metadata returns the resource type name.
func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal id of the group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the group",
			},
			"is_system": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is a system group",
				Default:     booldefault.StaticBool(false),
			},
			"redirect_on_login": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Path to redirect members to upon login",
				Default:     stringdefault.StaticString("/"),
			},
			"permissions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Global permissions for this group (see: https://github.com/requarks/wiki/blob/db8a09fe8c267a54fbbfabe0dc871a2108824968/client/components/admin/admin-groups-edit-permissions.vue#L43)",
			},
			"page_rules": schema.ListNestedAttribute{
				Required:    true,
				Description: "Page rules for this group. See nested object",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Internal ID for this rule",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"deny": schema.BoolAttribute{
							Required:    true,
							Description: "Defines whether this is a deny or allow rule",
						},
						"match": schema.StringAttribute{
							Required:            true,
							Description:         "Match pattern for this rule.",
							MarkdownDescription: "Match pattern for this rule. One of:\n  - START\n  - EXACT\n  - END\n  - REGEX\n  - TAG",
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(wikijs.PageRuleMatchStart),
									string(wikijs.PageRuleMatchExact),
									string(wikijs.PageRuleMatchEnd),
									string(wikijs.PageRuleMatchRegex),
									string(wikijs.PageRuleMatchTag),
								),
							},
						},
						"roles": schema.ListAttribute{
							Required:    true,
							Description: "Permissions of this role (see: https://github.com/requarks/wiki/blob/db8a09fe8c267a54fbbfabe0dc871a2108824968/client/components/admin/admin-groups-edit-permissions.vue#L43)",
							ElementType: types.StringType,
						},
						"path": schema.StringAttribute{
							Required:    true,
							Description: "Path to match on",
						},
						"locales": schema.ListAttribute{
							Required:    true,
							Description: "Locale to match on. Empty list to match on any locale.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)

	pageRules := make([]wikijs.PageRuleInput, len(data.PageRules))
	for i, r := range data.PageRules {
		var roles []string
		resp.Diagnostics.Append(r.Roles.ElementsAs(ctx, &roles, false)...)

		var locales []string
		resp.Diagnostics.Append(r.Locales.ElementsAs(ctx, &locales, false)...)

		pageRules[i] = wikijs.PageRuleInput{
			Id:      fmt.Sprintf("tf_%d_%d", data.Id.ValueInt64(), i),
			Deny:    r.Deny.ValueBool(),
			Match:   wikijs.PageRuleMatch(r.Match.ValueString()),
			Roles:   roles,
			Path:    r.Path.ValueString(),
			Locales: locales,
		}

		data.PageRules[i].Id = types.StringValue(pageRules[i].Id)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.CreateGroup(ctx, r.client.graphql, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Wiki.js Group Request failed", err.Error())
		return
	}
	if !wresp.Groups.Create.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not create Wiki.js group: %s", wresp.Groups.Create.ResponseResult.Slug), wresp.Groups.Create.ResponseResult.Message)
		return
	}

	data.Id = types.Int64Value(int64(wresp.Groups.Create.Group.Id))
	data.IsSystem = types.BoolValue(wresp.Groups.Create.Group.IsSystem)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)

	wresp2, err := wikijs.UpdateGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()), data.Name.ValueString(), data.RedirectOnLogin.ValueString(), permissions, pageRules)
	if err != nil {
		resp.Diagnostics.AddWarning("Finalize Wiki.js Group Request failed", err.Error())
		return
	}
	if !wresp2.Groups.Update.ResponseResult.Succeeded {
		resp.Diagnostics.AddWarning(fmt.Sprintf("Could not finalize Wiki.js group: %s", wresp2.Groups.Update.ResponseResult.Slug), wresp2.Groups.Update.ResponseResult.Message)
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Get Wiki.JS Group Request failed", err.Error())
		return
	}

	if wresp.Groups.Single.Id == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	if data.Id.ValueInt64() != int64(wresp.Groups.Single.Id) {
		resp.Diagnostics.AddError("Wiki.js returned the wrong group", fmt.Sprintf("expected group %d, got %d", data.Id.ValueInt64(), wresp.Groups.Single.Id))
		return
	}

	data.Name = types.StringValue(wresp.Groups.Single.Name)
	data.IsSystem = types.BoolValue(wresp.Groups.Single.IsSystem)
	data.RedirectOnLogin = types.StringValue(wresp.Groups.Single.RedirectOnLogin)

	permissions, diag := types.ListValueFrom(ctx, types.StringType, wresp.Groups.Single.Permissions)
	resp.Diagnostics = append(resp.Diagnostics, diag...)
	data.Permissions = permissions

	data.PageRules = []groupPageRuleModel{}
	for _, r := range wresp.Groups.Single.PageRules {
		roles, diag := types.ListValueFrom(ctx, types.StringType, r.Roles)
		resp.Diagnostics = append(resp.Diagnostics, diag...)
		locales, diag := types.ListValueFrom(ctx, types.StringType, r.Locales)
		resp.Diagnostics = append(resp.Diagnostics, diag...)

		data.PageRules = append(data.PageRules, groupPageRuleModel{
			Id:      types.StringValue(r.Id),
			Deny:    types.BoolValue(r.Deny),
			Match:   types.StringValue(string(r.Match)),
			Roles:   roles,
			Path:    types.StringValue(r.Path),
			Locales: locales,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)

	pageRules := make([]wikijs.PageRuleInput, len(data.PageRules))
	for i, r := range data.PageRules {
		var roles []string
		resp.Diagnostics.Append(r.Roles.ElementsAs(ctx, &roles, false)...)

		var locales []string
		resp.Diagnostics.Append(r.Locales.ElementsAs(ctx, &locales, false)...)

		pageRules[i] = wikijs.PageRuleInput{
			Id:      r.Id.ValueString(),
			Deny:    r.Deny.ValueBool(),
			Match:   wikijs.PageRuleMatch(r.Match.ValueString()),
			Roles:   roles,
			Path:    r.Path.ValueString(),
			Locales: locales,
		}
	}

	wresp, err := wikijs.UpdateGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()), data.Name.ValueString(), data.RedirectOnLogin.ValueString(), permissions, pageRules)
	if err != nil {
		resp.Diagnostics.AddError("Update Wiki.js Group Request failed", err.Error())
		return
	}
	if !wresp.Groups.Update.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update Wiki.js group: %s", wresp.Groups.Update.ResponseResult.Slug), wresp.Groups.Update.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform plan data into the model
	var data *groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.DeleteGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Delete Wiki.js Group Request failed", err.Error())
		return
	}
	if !wresp.Groups.Delete.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not delete Wiki.js group: %s", wresp.Groups.Delete.ResponseResult.Slug), wresp.Groups.Delete.ResponseResult.Message)
		return
	}
}
