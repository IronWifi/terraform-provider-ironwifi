package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
)

var _ datasource.DataSource = &AuthProvidersDataSource{}

// AuthProvidersDataSource lists IronWiFi authentication providers.
type AuthProvidersDataSource struct {
	client *client.Client
}

// AuthProvidersDataSourceModel is the schema model for the auth providers data source.
type AuthProvidersDataSourceModel struct {
	NameFilter types.String            `tfsdk:"name_filter"`
	TypeFilter types.String            `tfsdk:"type_filter"`
	Items      []AuthProviderItemModel `tfsdk:"items"`
}

// AuthProviderItemModel represents a single authentication provider in the data source output.
type AuthProviderItemModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Status          types.String `tfsdk:"status"`
	CaptivePortalID types.String `tfsdk:"captive_portal_id"`
}

// NewAuthProvidersDataSource returns a new authentication providers data source.
func NewAuthProvidersDataSource() datasource.DataSource {
	return &AuthProvidersDataSource{}
}

func (d *AuthProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authentication_providers"
}

func (d *AuthProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists IronWiFi authentication providers.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to provider names.",
				Optional:    true,
			},
			"type_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to provider type.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of authentication providers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true, Description: "Provider ID."},
						"name":              schema.StringAttribute{Computed: true, Description: "Provider name."},
						"type":              schema.StringAttribute{Computed: true, Description: "Provider type (e.g. LDAP, SAML)."},
						"status":            schema.StringAttribute{Computed: true, Description: "Provider status."},
						"captive_portal_id": schema.StringAttribute{Computed: true, Description: "Associated captive portal ID."},
					},
				},
			},
		},
	}
}

func (d *AuthProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *AuthProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AuthProvidersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.List("authentication-providers", "authentication_providers")
	if err != nil {
		resp.Diagnostics.AddError("Error listing authentication providers", err.Error())
		return
	}

	for _, item := range items {
		name := fmt.Sprintf("%v", item["name"])
		provType := fmt.Sprintf("%v", item["type"])

		if !state.NameFilter.IsNull() && !state.NameFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(state.NameFilter.ValueString())) {
				continue
			}
		}
		if !state.TypeFilter.IsNull() && !state.TypeFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(provType), strings.ToLower(state.TypeFilter.ValueString())) {
				continue
			}
		}

		state.Items = append(state.Items, AuthProviderItemModel{
			ID:              stringVal(item, "id"),
			Name:            types.StringValue(name),
			Type:            types.StringValue(provType),
			Status:          stringVal(item, "status"),
			CaptivePortalID: stringVal(item, "captive_portal_id"),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
