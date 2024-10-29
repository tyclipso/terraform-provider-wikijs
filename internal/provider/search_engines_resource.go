package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

var (
	_ resource.Resource                   = &searchEnginesResource{}
	_ resource.ResourceWithConfigure      = &searchEnginesResource{}
	_ resource.ResourceWithModifyPlan     = &searchEnginesResource{}
	_ resource.ResourceWithValidateConfig = &searchEnginesResource{}
)

func NewSearchEnginesResource() resource.Resource {
	return &searchEnginesResource{}
}

type searchEnginesResource struct {
	client *WikiJSClient
}

type searchEnginesResourceModel struct {
	SearchEngines []searchEnginesModel `tfsdk:"search_engines"`
}

type searchEnginesModel struct {
	IsEnabled types.Bool   `tfsdk:"is_enabled"`
	Key       types.String `tfsdk:"key"`
	Config    types.Map    `tfsdk:"config"`
}

func (r *searchEnginesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_search_engines"
}

func (r *searchEnginesResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"search_engines": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of search engines and their config.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"is_enabled": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Is the search engine enabled or not.\n",
						},
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The registered/machine name of the search engine used to identify it in the system.\n",
						},
						"config": schema.MapAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							Sensitive:           true,
							MarkdownDescription: "Map with config options for this specific search engine.\n",
						},
					},
				},
			},
		},
		MarkdownDescription: "The `{{ .Name }}` {{ .Type }} implements the WikiJS API mutatation `search{updateSearchEngines{engines{…}}}`.\n" +
			"\n" +
			"**Be aware**.\n" +
			"This {{ .Type }} supports only one instance as the implementation needs the complete configuration of search engines.\n" +
			"You cannot specify one search engine with its config as this will erase the complete list.\n" +
			"If you are unsure about the `key` and `config` fields, you can query the API after activating it manually or with the `wikijs_api` resource.\n" +
			"Make sure that the search provider `isAvailable` before you try to set `isEnabled` and that you have always exactly one search engine enabled.\n" +
			"The query with the minimal information you need is as follows:\n" +
			"\n" +
			"```graphql\n" +
			"query{\n" +
			"  search{\n" +
			"    searchEngines{\n" +
			"      title,\n" +
			"      key,\n" +
			"      isEnabled,\n" +
			"      isAvailable,\n" +
			"      config{\n" +
			"        key,\n" +
			"        value\n" +
			"      }\n" +
			"    }" +
			"  }\n" +
			"}\n" +
			"```\n" +
			"\n" +
			"The `title` field is not needed in the `{{ .Name }}` {{ .Type }} but is necessary for the graphql query.\n" +
			"You get back a JSON in the `config.value` field where you have to look for another `value` field.",
	}
}

func (r *searchEnginesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *searchEnginesResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var data, plan, state *searchEnginesResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !reflect.DeepEqual(plan, state) {
		resp.Diagnostics.AddWarning(
			"Resource is created or changed and will be updated",
			"This will trigger an IndexRebuild. "+
				"Please make sure that this additional load does not impact the performance.",
		)
	}

	wresp, err := wikijs.GetSearchEngines(ctx, r.client.graphql, "", "")
	if err != nil {
		resp.Diagnostics.AddError("Get Search Engines Request failed", err.Error())
	}
	isAvailable := map[string]bool{}
	for _, ws := range wresp.Search.SearchEngines {
		isAvailable[ws.Key] = ws.IsAvailable
	}

	for _, s := range data.SearchEngines {
		if s.IsEnabled.ValueBool() && !isAvailable[s.Key.ValueString()] {
			resp.Diagnostics.AddAttributeError(
				path.Root("is_enabled"),
				"Attribute Configured Wrong",
				fmt.Sprintf("%s cannot be enabled as search engine as it is not made available in the wikijs.", s.Key.ValueString()),
			)
		}
	}
}

func (r *searchEnginesResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data *searchEnginesResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	isEnabled := []string{}

	for _, s := range data.SearchEngines {
		if s.IsEnabled.ValueBool() {
			isEnabled = append(isEnabled, s.Key.ValueString())
		}
	}

	if len(isEnabled) < 1 {
		resp.Diagnostics.AddAttributeError(
			path.Root("is_enabled"),
			"Attribute Configured Wrong",
			"At any given time exactly one search engine needs to be enabled.\n"+
				"Maybe enable the default search engine with the key 'db'.",
		)
	}

	if len(isEnabled) > 1 {
		k := strings.Join(isEnabled, " and ")
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Attribute Configured Wrong",
			fmt.Sprintf("Only one search engine can be enabled at any time.\n"+
				"Currently %s are enabled.", k),
		)
	}
}

func (r *searchEnginesResource) buildApiDataModel(ctx context.Context, data *searchEnginesResourceModel) ([]wikijs.SearchEngineInput, diag.Diagnostics) {
	var result []wikijs.SearchEngineInput

	used := map[string]bool{}

	for _, s := range data.SearchEngines {
		if used[s.Key.ValueString()] {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Search Engine Key must be unique", "Every search engine in the list attribute 'search_engines' in your wikijs_search_engines resource needs to have a unique attribute key.")}
		}
		used[s.Key.ValueString()] = true

		var configTmp map[string]string
		if diag := s.Config.ElementsAs(ctx, &configTmp, false); diag.HasError() {
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

		result = append(result, wikijs.SearchEngineInput{
			IsEnabled: s.IsEnabled.ValueBool(),
			Key:       s.Key.ValueString(),
			Config:    config,
		})
	}

	return result, nil
}

func (r *searchEnginesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *searchEnginesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searchEngines, diag := r.buildApiDataModel(ctx, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetSearchEngines(ctx, r.client.graphql, searchEngines)
	if err != nil {
		resp.Diagnostics.AddError("Set Search Engines Request failed", err.Error())
		return
	}
	if !wresp.Search.UpdateSearchEngines.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update search engines: %s", wresp.Search.UpdateSearchEngines.ResponseResult.Slug), wresp.Search.UpdateSearchEngines.ResponseResult.Message)
		return
	}
	wresp2, err := wikijs.RebuildSearchIndex(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Rebuild Search Index Request failed", err.Error())
	}
	if !wresp2.Search.RebuildIndex.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not rebuild search index: %s", wresp2.Search.RebuildIndex.ResponseResult.Slug), wresp2.Search.RebuildIndex.ResponseResult.Message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *searchEnginesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *searchEnginesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetSearchEngines(ctx, r.client.graphql, "", "")
	if err != nil {
		resp.Diagnostics.AddError("Get Search Engines Request failed.", err.Error())
	}
	type configValue struct {
		Value any `json:"value"`
	}
	data.SearchEngines = data.SearchEngines[:0]
	for i, s := range wresp.Search.SearchEngines {
		if len(data.SearchEngines) <= i {
			data.SearchEngines = append(data.SearchEngines, make([]searchEnginesModel, i+1-len(data.SearchEngines))...)
		}

		configTmp := map[string]string{}
		for _, c := range s.Config {
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

		data.SearchEngines[i].IsEnabled = types.BoolValue(s.IsEnabled)
		data.SearchEngines[i].Key = types.StringValue(s.Key)
		data.SearchEngines[i].Config = config
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *searchEnginesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *searchEnginesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searchEngines, diag := r.buildApiDataModel(ctx, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetSearchEngines(ctx, r.client.graphql, searchEngines)
	if err != nil {
		resp.Diagnostics.AddError("Set Search Engines Request failed", err.Error())
		return
	}
	if !wresp.Search.UpdateSearchEngines.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update search engines: %s", wresp.Search.UpdateSearchEngines.ResponseResult.Slug), wresp.Search.UpdateSearchEngines.ResponseResult.Message)
		return
	}
	wresp2, err := wikijs.RebuildSearchIndex(ctx, r.client.graphql)
	if err != nil {
		resp.Diagnostics.AddError("Rebuild Search Index Request failed", err.Error())
	}
	if !wresp2.Search.RebuildIndex.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not rebuild search index: %s", wresp2.Search.RebuildIndex.ResponseResult.Slug), wresp2.Search.RebuildIndex.ResponseResult.Message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *searchEnginesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Not changing search engines", "Deleting the wikijs_search_engines resource just removes the resource from the terraform state. No wiki.js config is changed.")
}
