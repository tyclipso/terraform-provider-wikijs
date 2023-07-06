package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.startnext.org/sre/terraform/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &managedSystemGroupResource{}
	_ resource.ResourceWithConfigure = &managedSystemGroupResource{}
)

// NewManagedSystemGroupResource is a helper function to simplify the provider implementation.
func NewManagedSystemGroupResource() resource.Resource {
	return &managedSystemGroupResource{}
}

// managedSystemGroupResource is the resource implementation.
type managedSystemGroupResource struct {
	client *WikiJSClient
}

type managedSystemGroupResourceModel struct {
	Id              types.Int64          `tfsdk:"group_id"`
	Name            types.String         `tfsdk:"name"`
	RedirectOnLogin types.String         `tfsdk:"redirect_on_login"`
	Permissions     types.List           `tfsdk:"permissions"`
	PageRules       []groupPageRuleModel `tfsdk:"page_rules"`
}

// Metadata returns the resource type name.
func (r *managedSystemGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_system_group"
}

// Schema defines the schema for the resource.
func (r *managedSystemGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group_id": schema.Int64Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"redirect_on_login": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString("/"),
			},
			"permissions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"page_rules": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"deny": schema.BoolAttribute{
							Required: true,
						},
						"match": schema.StringAttribute{
							Required: true,
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
							ElementType: types.StringType,
						},
						"path": schema.StringAttribute{
							Required: true,
						},
						"locales": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *managedSystemGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *managedSystemGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *managedSystemGroupResourceModel
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
	}

	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Get Wiki.JS Manged Group Request failed", err.Error())
		return
	}

	if wresp.Groups.Single.Id == 0 {
		resp.Diagnostics.AddError("Invalid system group id", fmt.Sprintf("There is no system group with id %d in Wiki.js", data.Id.ValueInt64()))
		return
	}

	if !wresp.Groups.Single.IsSystem {
		resp.Diagnostics.AddError("Not a system group", fmt.Sprintf("The Wiki.js group with id %d is not a system group", data.Id.ValueInt64()))
		return
	}

	if data.Id.ValueInt64() != int64(wresp.Groups.Single.Id) {
		resp.Diagnostics.AddError("Wiki.js returned the wrong group", fmt.Sprintf("expected group %d, got %d", data.Id.ValueInt64(), wresp.Groups.Single.Id))
		return
	}
	data.Name = types.StringValue(wresp.Groups.Single.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)

	wresp2, err := wikijs.UpdateGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()), data.Name.ValueString(), data.RedirectOnLogin.ValueString(), permissions, pageRules)
	if err != nil {
		resp.Diagnostics.AddError("Update Wiki.js system Group Request failed", err.Error())
		return
	}
	if !wresp2.Groups.Update.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update Wiki.js system group: %s", wresp2.Groups.Update.ResponseResult.Slug), wresp2.Groups.Update.ResponseResult.Message)
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *managedSystemGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *managedSystemGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetGroup(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Get Wiki.JS system Group Request failed", err.Error())
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
func (r *managedSystemGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *managedSystemGroupResourceModel
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
		resp.Diagnostics.AddError("Update Wiki.js system Group Request failed", err.Error())
		return
	}
	if !wresp.Groups.Update.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update Wiki.js system group: %s", wresp.Groups.Update.ResponseResult.Slug), wresp.Groups.Update.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *managedSystemGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Can not delete system group", "Deleting the wikijs_managed_system_group resource just removes the resource from the terraform state. The system group in wiki.js is not changed.")
}
