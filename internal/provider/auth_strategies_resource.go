package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tyclipso/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &authStrategiesResource{}
	_ resource.ResourceWithConfigure = &authStrategiesResource{}
)

// NewAuthStrategiesResource is a helper function to simplify the provider implementation.
func NewAuthStrategiesResource() resource.Resource {
	return &authStrategiesResource{}
}

// authStrategiesResource is the resource implementation.
type authStrategiesResource struct {
	client *WikiJSClient
}

type authStrategiesResourceModel struct {
	Strategies []authStrategieModel `tfsdk:"strategies"`
}

type authStrategieModel struct {
	Key              types.String `tfsdk:"key"`
	StrategyKey      types.String `tfsdk:"strategy_key"`
	Config           types.Map    `tfsdk:"config"`
	DisplayName      types.String `tfsdk:"display_name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	SelfRegistration types.Bool   `tfsdk:"self_registration"`
	DomainWhitelist  types.List   `tfsdk:"domain_whitelist"`
	AutoEnrollGroups types.List   `tfsdk:"auto_enroll_groups"`
}

// Metadata returns the resource type name.
func (r *authStrategiesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_strategies"
}

// Schema defines the schema for the resource.
func (r *authStrategiesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"strategies": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Unique Key for this instance of the auth strategy. This resource can generate a unique key for you, but when you change the order of your auth strategies you have to explicitly set this key by yourself.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"strategy_key": schema.StringAttribute{
							Required:    true,
							Description: "Unique Identifier of the auth strategie to use (e.g. local, oauth2, oidc, keycloak, ldap)",
							MarkdownDescription: "Unique Identifier of the auth strategie to use (e.g. local, oauth2, oidc, keycloak, ldap)\n\n" +
								"```graphql\n" +
								"query GetAuthStrategies {\n  authentication {\n    strategies {\n      title\n      key\n      description\n    }\n  }\n}\n" +
								"```",
						},
						"config": schema.MapAttribute{
							Required:    true,
							ElementType: types.StringType,
							Sensitive:   true,
							Description: "Map with config options for this specifc auth strategie",
							MarkdownDescription: "Map with config options for this specifc auth strategie\n\n" +
								"```graphql\n" +
								"query GetAuthStrategies {\n  authentication {\n    strategies {\n      title\n      key\n      props {\n        key\n        value\n      }\n    }\n  }\n}\n" +
								"```",
						},
						"display_name": schema.StringAttribute{
							Required:    true,
							Description: "Name for this instance to be shown in the interface and at the login screen",
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether to enable this auth strategy instance",
							Default:     booldefault.StaticBool(true),
						},
						"self_registration": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Automatically create user accounts for people who successfully login via this auth strategie",
							Default:     booldefault.StaticBool(false),
							Validators: []validator.Bool{
								boolvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("domain_whitelist"), path.MatchRelative().AtParent().AtName("auto_enroll_groups")),
							},
						},
						"domain_whitelist": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "When self_registration is set to true, this list must contain the allowed domains",
							Validators: []validator.List{
								listvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("self_registration"), path.MatchRelative().AtParent().AtName("auto_enroll_groups")),
							},
						},
						"auto_enroll_groups": schema.ListAttribute{
							Optional:    true,
							ElementType: types.Int64Type,
							Description: "When self_registration is set to true, this list must contain the group ids the newly created account is added to.",
							Validators: []validator.List{
								listvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("self_registration"), path.MatchRelative().AtParent().AtName("domain_whitelist")),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (d *authStrategiesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *authStrategiesResource) buildApiDataModel(ctx context.Context, data *authStrategiesResourceModel) ([]wikijs.AuthenticationStrategyInput, diag.Diagnostics) {
	var result []wikijs.AuthenticationStrategyInput

	used := map[string]bool{}

	for i, s := range data.Strategies {
		if s.Key.IsUnknown() {
			if s.StrategyKey.ValueString() == "local" {
				data.Strategies[i].Key = types.StringValue("local")
			} else {
				key, err := uuid.NewRandom()
				if err != nil {
					return nil, diag.Diagnostics{diag.NewErrorDiagnostic("could not generate uuid", err.Error())}
				}
				data.Strategies[i].Key = types.StringValue(key.String())
			}
			s.Key = data.Strategies[i].Key
		}

		if used[s.Key.ValueString()] {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Auth Strategie Key must be unique", "Every auth strategy in the list attribute 'strategies' in you wikijs_auth_strategies resource needs to have a unique attribute key.\nThis error might also occur when you change the order of your auth strategies and have not set the keys explicitly.")}
		}
		used[s.Key.ValueString()] = true

		var domainWhitelist []string
		if diag := s.DomainWhitelist.ElementsAs(ctx, &domainWhitelist, false); diag.HasError() {
			return nil, diag
		}
		if domainWhitelist == nil {
			domainWhitelist = []string{}
		}

		var autoEnrollGroupsTmp []int64
		if diag := s.AutoEnrollGroups.ElementsAs(ctx, &autoEnrollGroupsTmp, false); diag.HasError() {
			return nil, diag
		}
		autoEnrollGroups := make([]int, len(autoEnrollGroupsTmp))
		for i, g := range autoEnrollGroupsTmp {
			autoEnrollGroups[i] = int(g)
		}

		var configTmp map[string]string
		if diag := s.Config.ElementsAs(ctx, &configTmp, false); diag.HasError() {
			return nil, diag
		}
		config := make([]wikijs.KeyValuePairInput, 0, len(configTmp))
		for k, v := range configTmp {
			valueTmp := map[string]string{"v": v}
			value, err := json.Marshal(valueTmp)
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("could not marshal json", err.Error())}
			}

			config = append(config, wikijs.KeyValuePairInput{
				Key:   k,
				Value: string(value),
			})
		}

		result = append(result, wikijs.AuthenticationStrategyInput{
			Key:              s.Key.ValueString(),
			StrategyKey:      s.StrategyKey.ValueString(),
			Config:           config,
			DisplayName:      s.DisplayName.ValueString(),
			Order:            i,
			IsEnabled:        s.Enabled.ValueBool(),
			SelfRegistration: s.SelfRegistration.ValueBool(),
			DomainWhitelist:  domainWhitelist,
			AutoEnrollGroups: autoEnrollGroups,
		})
	}

	return result, nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *authStrategiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *authStrategiesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	strategies, diag := r.buildApiDataModel(ctx, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetAuthStrategies(ctx, r.client.graphql, strategies)
	if err != nil {
		resp.Diagnostics.AddError("Set Authentication Strategies Request failed", err.Error())
		return
	}
	if !wresp.Authentication.UpdateStrategies.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update authentication strategies: %s", wresp.Authentication.UpdateStrategies.ResponseResult.Slug), wresp.Authentication.UpdateStrategies.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *authStrategiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform plan data into the model
	var data *authStrategiesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetAuthStrategies(ctx, r.client.graphql, false)
	if err != nil {
		resp.Diagnostics.AddError("Get Authentication Strategies Request failed", err.Error())
		return
	}

	type configValue struct {
		Value any `json:"value"`
	}

	data.Strategies = data.Strategies[:0]
	for _, s := range wresp.Authentication.ActiveStrategies {
		if len(data.Strategies) <= s.Order {
			data.Strategies = append(data.Strategies, make([]authStrategieModel, s.Order+1-len(data.Strategies))...)
		}

		configTmp := map[string]string{}
		for _, c := range s.Config {
			valueObj := configValue{}
			if err := json.Unmarshal([]byte(c.Value), &valueObj); err != nil {
				resp.Diagnostics.AddError("could not unmarshal json", err.Error())
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

		data.Strategies[s.Order].Key = types.StringValue(s.Key)
		data.Strategies[s.Order].StrategyKey = types.StringValue(s.Strategy.Key)
		data.Strategies[s.Order].Config = config
		data.Strategies[s.Order].DisplayName = types.StringValue(s.DisplayName)
		data.Strategies[s.Order].Enabled = types.BoolValue(s.IsEnabled)
		data.Strategies[s.Order].SelfRegistration = types.BoolValue(s.SelfRegistration)
		if s.SelfRegistration {
			domainWhitelist, diag := types.ListValueFrom(ctx, types.StringType, s.DomainWhitelist)
			resp.Diagnostics.Append(diag...)
			data.Strategies[s.Order].DomainWhitelist = domainWhitelist

			autoEnrollGroupsTmp := make([]int64, len(s.AutoEnrollGroups))
			for i, g := range s.AutoEnrollGroups {
				autoEnrollGroupsTmp[i] = int64(g)
			}
			autoEnrollGroups, diag := types.ListValueFrom(ctx, types.Int64Type, autoEnrollGroupsTmp)
			resp.Diagnostics.Append(diag...)
			data.Strategies[s.Order].AutoEnrollGroups = autoEnrollGroups
		} else {
			data.Strategies[s.Order].DomainWhitelist = types.ListNull(types.StringType)
			data.Strategies[s.Order].AutoEnrollGroups = types.ListNull(types.Int64Type)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *authStrategiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *authStrategiesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	strategies, diag := r.buildApiDataModel(ctx, data)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.SetAuthStrategies(ctx, r.client.graphql, strategies)
	if err != nil {
		resp.Diagnostics.AddError("Set Authentication Strategies Request failed", err.Error())
		return
	}
	if !wresp.Authentication.UpdateStrategies.ResponseResult.Succeeded {
		resp.Diagnostics.AddError(fmt.Sprintf("Could not update authentication strategies: %s", wresp.Authentication.UpdateStrategies.ResponseResult.Slug), wresp.Authentication.UpdateStrategies.ResponseResult.Message)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *authStrategiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Not changing auth strategies", "Deleting the wikijs_auth_strategies resource just removes the resource from the terraform state. No wiki.js config is changed.")
}
