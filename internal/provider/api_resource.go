package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.startnext.org/sre/terraform/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &apiResource{}
	_ resource.ResourceWithConfigure = &apiResource{}
)

// NewApiResource is a helper function to simplify the provider implementation.
func NewApiResource() resource.Resource {
	return &apiResource{}
}

// apiResource is the resource implementation.
type apiResource struct {
	client *WikiJSClient
}

type apiResourceModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// Metadata returns the resource type name.
func (r *apiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api"
}

// Schema defines the schema for the resource.
func (r *apiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *apiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *apiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *apiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetApiState(ctx, r.client.graphql, data.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Set API State Request failed", err.Error())
		return
	}
	if !wresp.Authentication.SetApiState.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not change api state: %s", wresp.Authentication.SetApiState.ResponseResult.Slug), wresp.Authentication.SetApiState.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *apiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *apiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetApiState(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get API State Request failed", err.Error())
		return
	}
	data.Enabled = types.BoolValue(wresp.Authentication.ApiState)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *apiResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *apiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetApiState(ctx, r.client.graphql, data.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Set API State Request failed", err.Error())
		return
	}
	if !wresp.Authentication.SetApiState.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not change api state: %s", wresp.Authentication.SetApiState.ResponseResult.Slug), wresp.Authentication.SetApiState.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *apiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform plan data into the model
	var data *apiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetApiState(ctx, r.client.graphql, false)
	if err != nil {
		resp.Diagnostics.AddError("Set API State Request failed", err.Error())
		return
	}
	if !wresp.Authentication.SetApiState.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not change api state: %s", wresp.Authentication.SetApiState.ResponseResult.Slug), wresp.Authentication.SetApiState.ResponseResult.Message)
		return
	}
	resp.Diagnostics.AddWarning("Wiki.JS API disabled", "Deleting the wikijs_api terraform resource disableds the wiki.js API as a security precaution.")
}
