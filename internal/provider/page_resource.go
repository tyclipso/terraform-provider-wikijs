package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &pageResource{}
	_ resource.ResourceWithConfigure   = &pageResource{}
	_ resource.ResourceWithImportState = &pageResource{}
)

// NewPageResource is a helper function to simplify the provider implementation.
func NewPageResource() resource.Resource {
	return &pageResource{}
}

// pageResource is the resource implementation.
type pageResource struct {
	client *WikiJSClient
}

type pageResourceModel struct {
	Id               types.Int64  `tfsdk:"id"`
	Path             types.String `tfsdk:"path"`
	Hash             types.String `tfsdk:"hash"`
	Title            types.String `tfsdk:"title"`
	Description      types.String `tfsdk:"description"`
	IsPrivate        types.Bool   `tfsdk:"is_private"`
	IsPublished      types.Bool   `tfsdk:"is_published"`
	PrivateNS        types.String `tfsdk:"private_ns"`
	PublishStartDate types.String `tfsdk:"publish_start_date"`
	PublishEndDate   types.String `tfsdk:"publish_end_date"`
	Tags             types.Set    `tfsdk:"tags"`
	Content          types.String `tfsdk:"content"`
	CreatedAt        types.String `tfsdk:"created_at"`
	Editor           types.String `tfsdk:"editor"`
	Locale           types.String `tfsdk:"locale"`
	ScriptCss        types.String `tfsdk:"script_css"`
	ScriptJs         types.String `tfsdk:"script_js"`
	CreatorId        types.Int64  `tfsdk:"creator_id"`
	CreatorName      types.String `tfsdk:"creator_name"`
	CreatorEmail     types.String `tfsdk:"creator_email"`
}

// Metadata returns the resource type name.
func (r *pageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page"
}

// Schema defines the schema for the resource.
func (r *pageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Internal id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Path of the page (omit leading slash)",
				PlanModifiers: []planmodifier.String{
					// @Todo: Implement move
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hash": schema.StringAttribute{
				Computed:    true,
				Description: "Page hash computed by wiki.js (see: https://github.com/requarks/wiki/blob/db8a09fe8c267a54fbbfabe0dc871a2108824968/server/helpers/page.js#L71)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Page Title",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Meta description of the page for search engines",
			},
			"is_private": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this is a private page",
				Default:     booldefault.StaticBool(false),
			},
			"is_published": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this page is published",
				Default:     booldefault.StaticBool(true),
			},
			"private_ns": schema.StringAttribute{
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"publish_start_date": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set to an RFC 3399 timestamp to define a publish date.",
				Default:     stringdefault.StaticString(""),
			},
			"publish_end_date": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set to an RFC 3399 timestamp to define an unpublish date.",
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of page tags",
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "Content of the page (format is defined by editor)",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Creation date of this page (expect RFC 3399 timestamp). Use data source to get updated_at",
			},
			"editor": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("markdown"),
				Description: "Editor type to use for this page",
			},
			"locale": schema.StringAttribute{
				Required:    true,
				Description: "Language of this page",
				PlanModifiers: []planmodifier.String{
					// @Todo: Implement move
					stringplanmodifier.RequiresReplace(),
				},
			},
			"script_css": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Additional CSS to add to the rendered page",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"script_js": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Additional JS to add to the rendered page",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creator_id": schema.Int64Attribute{
				Computed:    true,
				Description: "User id of the creator. Use data source to get authors",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"creator_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the page creator. Use data source to get authors",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creator_email": schema.StringAttribute{
				Computed:    true,
				Description: "Email of the page creator. Use data source to get authors",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *pageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *pageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *pageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	if data.Tags.IsNull() || data.Tags.IsUnknown() {
		tags = []string{}
		if data.Tags.IsUnknown() {
			data.Tags = types.SetNull(types.StringType)
		}
	} else {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	}

	wresp, err := wikijs.CreatePage(ctx, r.client.graphql,
		data.Content.ValueString(),
		data.Description.ValueString(),
		data.Editor.ValueString(),
		data.IsPublished.ValueBool(),
		data.IsPrivate.ValueBool(),
		data.Locale.ValueString(),
		data.Path.ValueString(),
		data.PublishEndDate.ValueString(),
		data.PublishStartDate.ValueString(),
		data.ScriptCss.ValueString(),
		data.ScriptJs.ValueString(),
		tags,
		data.Title.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Create Page Request failed", err.Error())
		return
	}
	if !wresp.Pages.Create.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not create page: %s", wresp.Pages.Create.ResponseResult.Slug), wresp.Pages.Create.ResponseResult.Message)
		return
	}

	data.Id = types.Int64Value(int64(wresp.Pages.Create.Page.Id))
	data.Hash = types.StringValue(wresp.Pages.Create.Page.Hash)
	data.PrivateNS = types.StringValue(wresp.Pages.Create.Page.PrivateNS)
	if data.PublishStartDate.IsUnknown() {
		data.PublishStartDate = types.StringValue(wresp.Pages.Create.Page.PublishStartDate)
	}
	if data.PublishEndDate.IsUnknown() {
		data.PublishEndDate = types.StringValue(wresp.Pages.Create.Page.PublishEndDate)
	}
	data.CreatedAt = types.StringValue(wresp.Pages.Create.Page.CreatedAt)
	data.ScriptCss = types.StringValue(wresp.Pages.Create.Page.ScriptCss)
	data.ScriptJs = types.StringValue(wresp.Pages.Create.Page.ScriptJs)
	data.CreatorId = types.Int64Value(int64(wresp.Pages.Create.Page.CreatorId))
	data.CreatorName = types.StringValue(wresp.Pages.Create.Page.CreatorName)
	data.CreatorEmail = types.StringValue(wresp.Pages.Create.Page.CreatorEmail)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *pageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *pageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetPage(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		if list, ok := err.(gqlerror.List); ok {
			for _, e := range list {
				if e.Message == "This page does not exist." {
					resp.State.RemoveResource(ctx)
					return
				}
			}
		}
		resp.Diagnostics.AddError("Read Page Request failed", err.Error())
		return
	}

	if wresp.Pages.Single.Id != int(data.Id.ValueInt64()) {
		resp.Diagnostics.AddError("Wiki.js returned the wrong page", fmt.Sprintf("expected page %d, got %d", data.Id.ValueInt64(), wresp.Pages.Single.Id))
		return
	}

	data.Path = types.StringValue(wresp.Pages.Single.Path)
	data.Hash = types.StringValue(wresp.Pages.Single.Hash)
	data.Title = types.StringValue(wresp.Pages.Single.Title)
	data.Description = types.StringValue(wresp.Pages.Single.Description)
	data.IsPrivate = types.BoolValue(wresp.Pages.Single.IsPrivate)
	data.IsPublished = types.BoolValue(wresp.Pages.Single.IsPublished)
	data.PrivateNS = types.StringValue(wresp.Pages.Single.PrivateNS)
	data.PublishStartDate = types.StringValue(wresp.Pages.Single.PublishStartDate)
	data.PublishEndDate = types.StringValue(wresp.Pages.Single.PublishEndDate)

	var tags []string
	for _, t := range wresp.Pages.Single.Tags {
		tags = append(tags, t.Tag)
	}
	if tags == nil && data.Tags.IsNull() {
		data.Tags = types.SetNull(types.StringType)
	} else {
		t, diag := types.SetValueFrom(ctx, types.StringType, tags)
		resp.Diagnostics.Append(diag...)
		data.Tags = t
	}

	data.Content = types.StringValue(wresp.Pages.Single.Content)
	data.CreatedAt = types.StringValue(wresp.Pages.Single.CreatedAt)
	data.Editor = types.StringValue(wresp.Pages.Single.Editor)
	data.Locale = types.StringValue(wresp.Pages.Single.Locale)
	data.ScriptCss = types.StringValue(wresp.Pages.Single.ScriptCss)
	data.ScriptJs = types.StringValue(wresp.Pages.Single.ScriptJs)
	data.CreatorId = types.Int64Value(int64(wresp.Pages.Single.CreatorId))
	data.CreatorName = types.StringValue(wresp.Pages.Single.CreatorName)
	data.CreatorEmail = types.StringValue(wresp.Pages.Single.CreatorEmail)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *pageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *pageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	if data.Tags.IsNull() || data.Tags.IsUnknown() {
		tags = []string{}
		if data.Tags.IsUnknown() {
			data.Tags = types.SetNull(types.StringType)
		}
	} else {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	}

	wresp, err := wikijs.UpdatePage(ctx, r.client.graphql,
		int(data.Id.ValueInt64()),
		data.Content.ValueString(),
		data.Description.ValueString(),
		data.Editor.ValueString(),
		data.IsPublished.ValueBool(),
		data.IsPrivate.ValueBool(),
		data.Locale.ValueString(),
		data.Path.ValueString(),
		data.PublishEndDate.ValueString(),
		data.PublishStartDate.ValueString(),
		data.ScriptCss.ValueString(),
		data.ScriptJs.ValueString(),
		tags,
		data.Title.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Update Page Request failed", err.Error())
		return
	}
	if !wresp.Pages.Update.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update page: %s", wresp.Pages.Update.ResponseResult.Slug), wresp.Pages.Update.ResponseResult.Message)
		return
	}

	if data.Hash.IsUnknown() {
		data.Hash = types.StringValue(wresp.Pages.Update.Page.Hash)
	}
	if data.PrivateNS.IsUnknown() {
		data.PrivateNS = types.StringValue(wresp.Pages.Update.Page.PrivateNS)
	}
	if data.PublishStartDate.IsUnknown() {
		data.PublishStartDate = types.StringValue(wresp.Pages.Update.Page.PublishStartDate)
	}
	if data.PublishEndDate.IsUnknown() {
		data.PublishEndDate = types.StringValue(wresp.Pages.Update.Page.PublishEndDate)
	}
	if data.CreatedAt.IsUnknown() {
		data.CreatedAt = types.StringValue(wresp.Pages.Update.Page.CreatedAt)
	}
	if data.ScriptCss.IsUnknown() {
		data.ScriptCss = types.StringValue(wresp.Pages.Update.Page.ScriptCss)
	}
	if data.ScriptJs.IsUnknown() {
		data.ScriptJs = types.StringValue(wresp.Pages.Update.Page.ScriptJs)
	}
	if data.CreatorId.IsUnknown() {
		data.CreatorId = types.Int64Value(int64(wresp.Pages.Update.Page.CreatorId))
	}
	if data.CreatorName.IsUnknown() {
		data.CreatorName = types.StringValue(wresp.Pages.Update.Page.CreatorName)
	}
	if data.CreatorEmail.IsUnknown() {
		data.CreatorEmail = types.StringValue(wresp.Pages.Update.Page.CreatorEmail)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *pageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform plan data into the model
	var data *pageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.DeletePage(ctx, r.client.graphql, int(data.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Delete Wiki.js Page Request failed", err.Error())
		return
	}
	if !wresp.Pages.Delete.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not delete Wiki.js page: %s", wresp.Pages.Delete.ResponseResult.Slug), wresp.Pages.Delete.ResponseResult.Message)
		return
	}
}

func (r *pageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if id, err := strconv.Atoi(req.ID); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("id"), "Could not parse id", err.Error())
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(id))...)
	}
}
