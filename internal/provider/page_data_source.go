package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &pageDataSource{}
	_ datasource.DataSourceWithConfigure = &pageDataSource{}
)

// NewPageDataSource is a helper function to simplify the provider implementation.
func NewPageDataSource() datasource.DataSource {
	return &pageDataSource{}
}

// pageDataSource is the data source implementation.
type pageDataSource struct {
	client *WikiJSClient
}

// pageDataSourceModel maps the data source schema data.
type pageDataSourceModel struct {
	Id               types.Int64    `tfsdk:"page_id"`
	Path             types.String   `tfsdk:"path"`
	Hash             types.String   `tfsdk:"hash"`
	Title            types.String   `tfsdk:"title"`
	Description      types.String   `tfsdk:"description"`
	IsPrivate        types.Bool     `tfsdk:"is_private"`
	IsPublished      types.Bool     `tfsdk:"is_published"`
	PrivateNS        types.String   `tfsdk:"private_ns"`
	PublishStartDate types.String   `tfsdk:"publish_start_date"`
	PublishEndDate   types.String   `tfsdk:"publish_end_date"`
	Tags             []pageTagModel `tfsdk:"tags"`
	Content          types.String   `tfsdk:"content"`
	Render           types.String   `tfsdk:"render"`
	ContentType      types.String   `tfsdk:"content_type"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	UpdatedAt        types.String   `tfsdk:"updated_at"`
	Editor           types.String   `tfsdk:"editor"`
	Locale           types.String   `tfsdk:"locale"`
	ScriptCss        types.String   `tfsdk:"script_css"`
	ScriptJs         types.String   `tfsdk:"script_js"`
	AuthorId         types.Int64    `tfsdk:"author_id"`
	AuthorName       types.String   `tfsdk:"author_name"`
	AuthorEmail      types.String   `tfsdk:"author_email"`
	CreatorId        types.Int64    `tfsdk:"creator_id"`
	CreatorName      types.String   `tfsdk:"creator_name"`
	CreatorEmail     types.String   `tfsdk:"creator_email"`
}

type pageTagModel struct {
	Id        types.Int64  `tfsdk:"id"`
	Tag       types.String `tfsdk:"tag"`
	Title     types.String `tfsdk:"title"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *pageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page"
}

// Schema defines the schema for the data source.
func (d *pageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"page_id": schema.Int64Attribute{
				Computed:    true,
				Optional:    true,
				Description: "Internal ID of this page",
				Validators: []validator.Int64{
					int64validator.AtLeastOneOf(path.MatchRoot("path")),
					int64validator.ConflictsWith(path.MatchRoot("locale")),
				},
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Path of the page (omit leading slash)",
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("id")),
					stringvalidator.AlsoRequires(path.MatchRoot("locale")),
				},
			},
			"hash": schema.StringAttribute{
				Computed:    true,
				Description: "Page hash computed by wiki.js (see: https://github.com/requarks/wiki/blob/db8a09fe8c267a54fbbfabe0dc871a2108824968/server/helpers/page.js#L71)",
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: "Page Title",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Meta description of the page for search engines",
			},
			"is_private": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is a private page",
			},
			"is_published": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this page is published",
			},
			"private_ns": schema.StringAttribute{
				Computed: true,
			},
			"publish_start_date": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3399 timestamp, when a publish date is defined.",
			},
			"publish_end_date": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3399 timestamp, when an unpublish date is defined.",
			},
			"tags": schema.SetNestedAttribute{
				Computed:    true,
				Description: "List of page tags",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Internal id of this tag",
						},
						"tag": schema.StringAttribute{
							Computed:    true,
							Description: "The actual tag name. Use this string, when referencing a tag",
						},
						"title": schema.StringAttribute{
							Computed:    true,
							Description: "Display name of this tag",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "Creation date of this tag (expect RFC 3399 timestamp)",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "Update date of this tag (expect RFC 3399 timestamp)",
						},
					},
				},
			},
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "Content of the page (format is defined by editor)",
			},
			"render": schema.StringAttribute{
				Computed:    true,
				Description: "Rendered HTML of the content",
			},
			"content_type": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation date of this page (expect RFC 3399 timestamp)",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Update date of this page (expect RFC 3399 timestamp)",
			},
			"editor": schema.StringAttribute{
				Computed:    true,
				Description: "Editor type of this page",
			},
			"locale": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Language of this page",
			},
			"script_css": schema.StringAttribute{
				Computed:    true,
				Description: "Additional CSS to add to the rendered page",
			},
			"script_js": schema.StringAttribute{
				Computed:    true,
				Description: "Additional JS to add to the rendered page",
			},
			"author_id": schema.Int64Attribute{
				Description: "User id of the last author.",
				Computed:    true,
			},
			"author_name": schema.StringAttribute{
				Description: "Name of the page last author.",
				Computed:    true,
			},
			"author_email": schema.StringAttribute{
				Description: "Email of the page last author.",
				Computed:    true,
			},
			"creator_id": schema.Int64Attribute{
				Description: "User id of the creator.",
				Computed:    true,
			},
			"creator_name": schema.StringAttribute{
				Description: "Name of the page creator.",
				Computed:    true,
			},
			"creator_email": schema.StringAttribute{
				Description: "Email of the page creator.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *pageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *pageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	type Page interface {
		GetId() int
		GetPath() string
		GetHash() string
		GetTitle() string
		GetDescription() string
		GetIsPrivate() bool
		GetIsPublished() bool
		GetPrivateNS() string
		GetPublishStartDate() string
		GetPublishEndDate() string
		GetContent() string
		GetRender() string
		GetContentType() string
		GetCreatedAt() string
		GetUpdatedAt() string
		GetEditor() string
		GetLocale() string
		GetScriptCss() string
		GetScriptJs() string
		GetAuthorId() int
		GetAuthorName() string
		GetAuthorEmail() string
		GetCreatorId() int
		GetCreatorName() string
		GetCreatorEmail() string
	}

	var (
		page Page
		tags []pageTagModel
	)
	if !state.Id.IsNull() {
		wresp, err := wikijs.GetPage(ctx, d.client.graphql, int(state.Id.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError("Get Page Query failed", err.Error())
			return
		}
		page = &wresp.Pages.Single
		for _, t := range wresp.Pages.Single.Tags {
			tags = append(tags, pageTagModel{
				Id:        types.Int64Value(int64(t.Id)),
				Tag:       types.StringValue(t.Tag),
				Title:     types.StringValue(t.Title),
				CreatedAt: types.StringValue(t.CreatedAt),
				UpdatedAt: types.StringValue(t.UpdatedAt),
			})
		}
	} else {
		wresp, err := wikijs.GetPageByPath(ctx, d.client.graphql, state.Path.ValueString(), state.Locale.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Get Page by Path Query failed", err.Error())
			return
		}
		page = &wresp.Pages.SingleByPath
		for _, t := range wresp.Pages.SingleByPath.Tags {
			tags = append(tags, pageTagModel{
				Id:        types.Int64Value(int64(t.Id)),
				Tag:       types.StringValue(t.Tag),
				Title:     types.StringValue(t.Title),
				CreatedAt: types.StringValue(t.CreatedAt),
				UpdatedAt: types.StringValue(t.UpdatedAt),
			})
		}
	}

	state.Id = types.Int64Value(int64(page.GetId()))
	state.Path = types.StringValue(page.GetPath())
	state.Hash = types.StringValue(page.GetHash())
	state.Title = types.StringValue(page.GetTitle())
	state.Description = types.StringValue(page.GetDescription())
	state.IsPrivate = types.BoolValue(page.GetIsPrivate())
	state.IsPublished = types.BoolValue(page.GetIsPublished())
	state.PrivateNS = types.StringValue(page.GetPrivateNS())
	state.PublishStartDate = types.StringValue(page.GetPublishStartDate())
	state.PublishEndDate = types.StringValue(page.GetPublishEndDate())
	state.Tags = tags
	state.Content = types.StringValue(page.GetContent())
	state.Render = types.StringValue(page.GetRender())
	state.ContentType = types.StringValue(page.GetContentType())
	state.CreatedAt = types.StringValue(page.GetCreatedAt())
	state.UpdatedAt = types.StringValue(page.GetUpdatedAt())
	state.Editor = types.StringValue(page.GetEditor())
	state.Locale = types.StringValue(page.GetLocale())
	state.ScriptCss = types.StringValue(page.GetScriptCss())
	state.ScriptJs = types.StringValue(page.GetScriptJs())
	state.AuthorId = types.Int64Value(int64(page.GetAuthorId()))
	state.AuthorName = types.StringValue(page.GetAuthorName())
	state.AuthorEmail = types.StringValue(page.GetAuthorEmail())
	state.CreatorId = types.Int64Value(int64(page.GetCreatorId()))
	state.CreatorName = types.StringValue(page.GetCreatorName())
	state.CreatorEmail = types.StringValue(page.GetCreatorEmail())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
