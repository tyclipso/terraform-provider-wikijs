package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.startnext.org/sre/terraform/terraform-provider-wikijs/wikijs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &authStrategiesDataSource{}
	_ datasource.DataSourceWithConfigure = &authStrategiesDataSource{}
)

// NewAuthStrategiesDataSource is a helper function to simplify the provider implementation.
func NewAuthStrategiesDataSource() datasource.DataSource {
	return &authStrategiesDataSource{}
}

// authStrategiesDataSource is the data source implementation.
type authStrategiesDataSource struct {
	client *WikiJSClient
}

// authStrategiesDataSourceModel maps the data source schema data.
type authStrategiesDataSourceModel struct {
	Strategies []authStrategieModel `tfsdk:"strategies"`
}

// Metadata returns the data source type name.
func (d *authStrategiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_strategies"
}

// Schema defines the schema for the data source.
func (d *authStrategiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"strategies": schema.ListNestedAttribute{
				Required:    true,
				Description: "List of active auth strategies (active means configured, not necessary enabled)",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "Unique Key for this instance of the auth strategie",
						},
						"strategy_key": schema.StringAttribute{
							Computed:    true,
							Description: "Unique Identifier of the auth strategie to use (e.g. local, oauth2, oidc, keycloak, ldap)",
							MarkdownDescription: "Unique Identifier of the auth strategie to use (e.g. local, oauth2, oidc, keycloak, ldap)\n\n" +
								"```graphql\n" +
								"query GetAuthStrategies {\n  authentication {\n    strategies {\n      title\n      key\n      description\n    }\n  }\n}\n" +
								"```",
						},
						"config": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Sensitive:   true,
							Description: "Map with config options for this specifc auth strategie",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "Name for this instance to be shown in the interface and at the login screen",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether to enable this auth strategy instance",
						},
						"self_registration": schema.BoolAttribute{
							Computed:    true,
							Description: "Allow users to create a user account via this auth strategy",
						},
						"domain_whitelist": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "When self_registration is set to true, this list contains the allowed domains",
						},
						"auto_enroll_groups": schema.ListAttribute{
							Computed:    true,
							ElementType: types.Int64Type,
							Description: "When self_registration is set to true, this list contains the group ids the newly created account is added to.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *authStrategiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *authStrategiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state authStrategiesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wresp, err := wikijs.GetAuthStrategies(ctx, d.client.graphql, false)
	if err != nil {
		resp.Diagnostics.AddError("Get Authentication Strategies Request failed", err.Error())
		return
	}

	type configValue struct {
		Value any `json:"value"`
	}

	state.Strategies = state.Strategies[:0]
	for _, s := range wresp.Authentication.ActiveStrategies {
		if len(state.Strategies) <= s.Order {
			state.Strategies = append(state.Strategies, make([]authStrategieModel, s.Order+1-len(state.Strategies))...)
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

		state.Strategies[s.Order].Key = types.StringValue(s.Key)
		state.Strategies[s.Order].StrategyKey = types.StringValue(s.Strategy.Key)
		state.Strategies[s.Order].Config = config
		state.Strategies[s.Order].DisplayName = types.StringValue(s.DisplayName)
		state.Strategies[s.Order].Enabled = types.BoolValue(s.IsEnabled)
		state.Strategies[s.Order].SelfRegistration = types.BoolValue(s.SelfRegistration)
		if s.SelfRegistration {
			domainWhitelist, diag := types.ListValueFrom(ctx, types.StringType, s.DomainWhitelist)
			resp.Diagnostics.Append(diag...)
			state.Strategies[s.Order].DomainWhitelist = domainWhitelist

			autoEnrollGroupsTmp := make([]int64, len(s.AutoEnrollGroups))
			for i, g := range s.AutoEnrollGroups {
				autoEnrollGroupsTmp[i] = int64(g)
			}
			autoEnrollGroups, diag := types.ListValueFrom(ctx, types.Int64Type, autoEnrollGroupsTmp)
			resp.Diagnostics.Append(diag...)
			state.Strategies[s.Order].AutoEnrollGroups = autoEnrollGroups
		} else {
			state.Strategies[s.Order].DomainWhitelist = types.ListNull(types.StringType)
			state.Strategies[s.Order].AutoEnrollGroups = types.ListNull(types.Int64Type)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
