package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &localizationResource{}
	_ resource.ResourceWithConfigure = &localizationResource{}
)

// NewLocalizationResource is a helper function to simplify the provider implementation.
func NewLocalizationResource() resource.Resource {
	return &localizationResource{}
}

// localizationResource is the resource implementation.
type localizationResource struct {
	client *WikiJSClient
}

type localizationResourceModel struct {
	Locale      types.String `tfsdk:"locale"`
	AutoUpdate  types.Bool   `tfsdk:"auto_update"`
	Namespacing types.Bool   `tfsdk:"namespacing"`
	Namespaces  types.Set    `tfsdk:"namespaces"`
}

// Metadata returns the resource type name.
func (r *localizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_localization"
}

// Schema defines the schema for the resource.
func (r *localizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"locale": schema.StringAttribute{
				Required:    true,
				Description: "All UI text elements will be displayed in selected language.",
			},
			"auto_update": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Automatically download updates to this locale as they become available.",
			},
			"namespacing": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enables multiple language versions of the same page.",
			},
			"namespaces": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Set of locales enabled for multilingual namespacing. The base locale defined above must always be included.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *localizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *localizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *localizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var namespaces []string
	resp.Diagnostics.Append(data.Namespaces.ElementsAs(ctx, &namespaces, false)...)

	install := namespaces
	install = append(install, data.Locale.ValueString())
	for _, l := range install {
		wresp, err := wikijs.DownloadLocale(ctx, r.client.graphql, l)
		if err != nil {
			resp.Diagnostics.AddError("Download Locale Request failed", err.Error())
			continue
		}
		if !wresp.Localization.DownloadLocale.ResponseResult.Succeeded {
			resp.Diagnostics.AddError(fmt.Sprintf("Could not install locale '%s': %s", l, wresp.Localization.DownloadLocale.ResponseResult.Slug), wresp.Localization.DownloadLocale.ResponseResult.Message)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetLocalization(ctx, r.client.graphql, data.Locale.ValueString(), data.AutoUpdate.ValueBool(), data.Namespacing.ValueBool(), namespaces)
	if err != nil {
		resp.Diagnostics.AddError("Update Localization Request failed", err.Error())
		return
	}
	if !wresp.Localization.UpdateLocale.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update localization: %s", wresp.Localization.UpdateLocale.ResponseResult.Slug), wresp.Localization.UpdateLocale.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *localizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *localizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetLocalization(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Getting localization request failed", err.Error())
		return
	}

	data.Locale = types.StringValue(wresp.Localization.Config.Locale)
	data.AutoUpdate = types.BoolValue(wresp.Localization.Config.AutoUpdate)
	data.Namespacing = types.BoolValue(wresp.Localization.Config.Namespacing)
	namespaces, diag := types.SetValueFrom(ctx, types.StringType, wresp.Localization.Config.Namespaces)
	resp.Diagnostics.Append(diag...)
	data.Namespaces = namespaces

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *localizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *localizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var namespaces []string
	resp.Diagnostics.Append(data.Namespaces.ElementsAs(ctx, &namespaces, false)...)

	install := namespaces
	install = append(install, data.Locale.ValueString())
	for _, l := range install {
		wresp, err := wikijs.DownloadLocale(ctx, r.client.graphql, l)
		if err != nil {
			resp.Diagnostics.AddError("Download Locale Request failed", err.Error())
			continue
		}
		if !wresp.Localization.DownloadLocale.ResponseResult.Succeeded {
			resp.Diagnostics.AddError(fmt.Sprintf("Could not install locale '%s': %s", l, wresp.Localization.DownloadLocale.ResponseResult.Slug), wresp.Localization.DownloadLocale.ResponseResult.Message)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetLocalization(ctx, r.client.graphql, data.Locale.ValueString(), data.AutoUpdate.ValueBool(), data.Namespacing.ValueBool(), namespaces)
	if err != nil {
		resp.Diagnostics.AddError("Update Localization Request failed", err.Error())
		return
	}
	if !wresp.Localization.UpdateLocale.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update localization: %s", wresp.Localization.UpdateLocale.ResponseResult.Slug), wresp.Localization.UpdateLocale.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *localizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Localization has no factory default", "Deleting the wikijs_localization resource just removes the resource from the terraform state. The settings in wiki.js are not changed.")
}
