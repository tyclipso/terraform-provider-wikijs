package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.startnext.org/sre/terraform/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &apiKeyResource{}
	_ resource.ResourceWithConfigure = &apiKeyResource{}
)

// NewApiKeyResource is a helper function to simplify the provider implementation.
func NewApiKeyResource() resource.Resource {
	return &apiKeyResource{}
}

// apiKeyResource is the resource implementation.
type apiKeyResource struct {
	client *WikiJSClient
}

type apiKeyResourceModel struct {
	Id                   types.Int64  `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	FullAccess           types.Bool   `tfsdk:"full_access"`
	GroupId              types.Int64  `tfsdk:"group_id"`
	KeyShort             types.String `tfsdk:"key_short"`
	ExpiresIn            types.String `tfsdk:"expires_in"`
	Expiration           types.String `tfsdk:"expiration"`
	MinRemainingDuration types.String `tfsdk:"min_remaining_duration"`
	CreatedAt            types.String `tfsdk:"created_at"`
	Key                  types.String `tfsdk:"key"`
}

// Metadata returns the resource type name.
func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// Schema defines the schema for the resource.
func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"full_access": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Bool{
					boolvalidator.ExactlyOneOf(path.MatchRoot("full_access"), path.MatchRoot("group_id")),
				},
			},
			"group_id": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRoot("full_access"), path.MatchRoot("group_id")),
				},
			},
			"key_short": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_in": schema.StringAttribute{
				Required:    true,
				Description: "When creating an API Key wiki.js expects an expiration timespan (e. g. '1d' '20h')",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration": schema.StringAttribute{
				Computed:    true,
				Description: "The actual expiration date of this key",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"min_remaining_duration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set a minimum duration the api key needs to remain active. This field is changed when the expiration dates comes to close and triggers a replace",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *apiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.CreateApiKey(ctx, r.client.graphql, data.Name.ValueString(), data.ExpiresIn.ValueString(), data.FullAccess.ValueBool(), int(data.GroupId.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Create API Key Request failed", err.Error())
		return
	}
	if !wresp.Authentication.CreateApiKey.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not create wiki.js API Key: %s", wresp.Authentication.CreateApiKey.ResponseResult.Slug), wresp.Authentication.CreateApiKey.ResponseResult.Message)
		return
	}
	data.Key = types.StringValue(wresp.Authentication.CreateApiKey.Key)

	wresp2, err := wikijs.GetApiKeys(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get API Keys Request failed", err.Error())
		return
	}
	for _, k := range wresp2.Authentication.ApiKeys {
		suffix, _ := strings.CutPrefix(k.KeyShort, "...")
		if k.Name == data.Name.ValueString() && strings.HasSuffix(wresp.Authentication.CreateApiKey.Key, suffix) {
			data.Id = types.Int64Value(int64(k.Id))
			data.KeyShort = types.StringValue(k.KeyShort)
			data.Expiration = types.StringValue(k.Expiration)
			data.CreatedAt = types.StringValue(k.CreatedAt)
			if data.MinRemainingDuration.IsUnknown() {
				data.MinRemainingDuration = types.StringNull()
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
			return
		}
	}

	resp.Diagnostics.AddError("Could not find API Key", "We successfully created an Wiki.JS API Key but could find it in the list of api keys to retrieve the id")
}

// Read refreshes the Terraform state with the latest data.
func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetApiKeys(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Get API Keys Request failed", err.Error())
		return
	}
	for _, k := range wresp.Authentication.ApiKeys {
		if k.Id == int(data.Id.ValueInt64()) {
			if k.IsRevoked {
				resp.State.RemoveResource(ctx)
				return
			}
			data.Name = types.StringValue(k.Name)
			data.KeyShort = types.StringValue(k.KeyShort)
			data.Expiration = types.StringValue(k.Expiration)
			if !data.MinRemainingDuration.IsNull() {
				minDur, err := time.ParseDuration(data.MinRemainingDuration.ValueString())
				if err != nil {
					resp.Diagnostics.AddAttributeWarning(path.Root("min_remaining_duration"), "Could not parse duration", err.Error())
				} else {
					expiration, err := time.Parse(time.RFC3339, k.Expiration)
					if err != nil {
						resp.Diagnostics.AddAttributeWarning(path.Root("min_remaining_duration"), "Could not parse expiration date", err.Error())
					} else {
						remaining := time.Until(expiration)
						if remaining < minDur {
							data.MinRemainingDuration = types.StringValue(remaining.String())
						}
					}
				}
			}
			data.CreatedAt = types.StringValue(k.CreatedAt)
			resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *apiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform plan data into the model
	var data *apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.RevokeApiKey(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Delete Wiki.js API Key Request failed", err.Error())
		return
	}
	if !wresp.Authentication.RevokeApiKey.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not revoke Wiki.js api key: %s", wresp.Authentication.RevokeApiKey.ResponseResult.Slug), wresp.Authentication.RevokeApiKey.ResponseResult.Message)
		return
	}
}
