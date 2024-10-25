package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

var (
	_ resource.Resource              = &renderersResource{}
	_ resource.ResourceWithConfigure = &renderersResource{}
)

func NewRenderersResource() resource.Resource {
	return &renderersResource{}
}

type renderersResource struct {
	client *WikiJSClient
}

type renderersResourceModel struct {
	Renderers []renderersModel `tfsdk:"renderers"`
}

type renderersModel struct {
	IsEnabled types.Bool   `tfsdk:"is_enabled"`
	Key       types.String `tfsdk:"key"`
	Config    types.Map    `tfsdk:"config"`
}

func (r *renderersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_renderers"
}

func (r *renderersResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"renderers": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"is_enabled": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Is the renderer enabled or not.",
						},
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The registered/machine name of the renderer used to identify it in the system.",
						},
						"config": schema.MapAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Map with config options for this specific renderer.\n",
						},
					},
				},
			},
		},
		MarkdownDescription: "The `{{ .Name }}` {{ .Type }} implements the WikiJS API mutatation `rendering{updateRenderers{renderers{…}}}`.\n" +
			"\n" +
			"**Be aware**.\n" +
			"This {{ .Type }} supports only one instance as the implementation needs the complete configuration of renderers.\n" +
			"You cannot specify one renderer with its config as this will erase the complete list.\n" +
			"If you are unsure about the `key` and `config` fields, you can query the API after activating it manually or with the `wikijs_api` resource.\n" +
			"The query with the minimal information you need is as follows:\n" +
			"\n" +
			"```graphql\n" +
			"query{\n" +
			"  rendering{\n" +
			"    renderers{\n" +
			"      title,\n" +
			"      key,\n" +
			"      isEnabled,\n" +
			"      config{\n" +
			"        key,\n" +
			"        value\n" +
			"      }\n" +
			"    }\n" +
			"  }\n" +
			"}\n" +
			"```\n" +
			"\n" +
			"The `title` field is not needed in the `{{ .Name }}` {{ .Type }} but is necessary for the graphql query.\n" +
			"You get back a JSON in the `config.value` field where you have to look for another `value` field.",
	}
}

func (r *renderersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*WikiJSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expexted *WikiJSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *renderersResource) buildApiDataModel(ctx context.Context, data *renderersResourceModel) ([]wikijs.RendererInput, diag.Diagnostics) {
	var result []wikijs.RendererInput

	used := map[string]bool{}

	for _, r := range data.Renderers {
		if used[r.Key.ValueString()] {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Renderer Key must be unique", "Every renderer in the list attribute 'renderers' in your wikijs_renderers resource needs to have a unique attribute key.")}
		}
		used[r.Key.ValueString()] = true

		var configTmp map[string]string
		if diag := r.Config.ElementsAs(ctx, &configTmp, false); diag.HasError() {
			return nil, diag
		}
		config := make([]wikijs.KeyValuePairInput, 0, len(configTmp))
		for k, v := range configTmp {
			valueTmp := map[string]string{"v": v}
			value, err := json.Marshal(valueTmp)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Could not marshal json", err.Error())}
			}

			config = append(config, wikijs.KeyValuePairInput{
				Key:   k,
				Value: string(value),
			})
		}

		result = append(result, wikijs.RendererInput{
			IsEnabled: r.IsEnabled.ValueBool(),
			Key:       r.Key.ValueString(),
			Config:    config,
		})
	}

	return result, nil
}

func (r *renderersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *renderersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	renderers, diag := r.buildApiDataModel(ctx, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetRenderers(ctx, r.client.graphql, renderers)
	if err != nil {
		resp.Diagnostics.AddError("Set Renderers Request failed", err.Error())
		return
	}
	if !wresp.Rendering.UpdateRenderers.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update renderers: %s", wresp.Rendering.UpdateRenderers.ResponseResult.Slug), wresp.Rendering.UpdateRenderers.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *renderersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *renderersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetRenderers(ctx, r.client.graphql, "", "")
	if err != nil {
		resp.Diagnostics.AddError("Get Renderers Request failed.", err.Error())
	}
	type configValue struct {
		Value any `json:"value"`
	}
	data.Renderers = data.Renderers[:0]
	for i, r := range wresp.Rendering.Renderers {
		if len(data.Renderers) <= i {
			data.Renderers = append(data.Renderers, make([]renderersModel, i+1-len(data.Renderers))...)
		}

		configTmp := map[string]string{}
		for _, c := range r.Config {
			valueObj := configValue{}
			if err := json.Unmarshal([]byte(c.Value), &valueObj); err != nil {
				resp.Diagnostics.AddError("Could not unmarshal JSON", err.Error())
				return
			}
			switch value := valueObj.Value.(type) {
			case bool:
				configTmp[c.Key] = strconv.FormatBool(value)
			case json.Number:
				configTmp[c.Key] = value.String()
			case string:
				configTmp[c.Key] = value
			default:
				configTmp[c.Key] = fmt.Sprintf("%v", value)
			}
		}

		config, diag := types.MapValueFrom(ctx, types.StringType, configTmp)
		resp.Diagnostics.Append(diag...)

		data.Renderers[i].IsEnabled = types.BoolValue(r.IsEnabled)
		data.Renderers[i].Key = types.StringValue(r.Key)
		data.Renderers[i].Config = config
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *renderersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *renderersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	renderers, diag := r.buildApiDataModel(ctx, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetRenderers(ctx, r.client.graphql, renderers)
	if err != nil {
		resp.Diagnostics.AddError("Set Renderers Request failed", err.Error())
		return
	}
	if !wresp.Rendering.UpdateRenderers.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update renderers: %s", wresp.Rendering.UpdateRenderers.ResponseResult.Slug), wresp.Rendering.UpdateRenderers.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *renderersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Not changing renderers", "Deleting the wikijs_renderers resource just removes the resource from the terraform state. No wiki.js config is changed.")
}
