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

var _ datasource.DataSource = &NetworksDataSource{}

// NetworksDataSource lists IronWiFi networks.
type NetworksDataSource struct {
	client *client.Client
}

// NetworksDataSourceModel is the schema model for the networks data source.
type NetworksDataSourceModel struct {
	NameFilter types.String        `tfsdk:"name_filter"`
	Items      []NetworkItemModel  `tfsdk:"items"`
}

// NetworkItemModel represents a single network in the data source output.
type NetworkItemModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Region    types.String `tfsdk:"region"`
	AuthPort  types.Int64  `tfsdk:"auth_port"`
	AcctPort  types.Int64  `tfsdk:"acct_port"`
	PrimaryIP types.String `tfsdk:"primary_ip"`
	Secret    types.String `tfsdk:"secret"`
}

// NewNetworksDataSource returns a new networks data source.
func NewNetworksDataSource() datasource.DataSource {
	return &NetworksDataSource{}
}

func (d *NetworksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *NetworksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists IronWiFi networks (RADIUS NAS clients).",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to network names.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of networks.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, Description: "Network ID."},
						"name":       schema.StringAttribute{Computed: true, Description: "Network name (nasname)."},
						"region":     schema.StringAttribute{Computed: true, Description: "Region."},
						"auth_port":  schema.Int64Attribute{Computed: true, Description: "RADIUS authentication port."},
						"acct_port":  schema.Int64Attribute{Computed: true, Description: "RADIUS accounting port."},
						"primary_ip": schema.StringAttribute{Computed: true, Description: "Primary IP address."},
						"secret":     schema.StringAttribute{Computed: true, Description: "RADIUS shared secret.", Sensitive: true},
					},
				},
			},
		},
	}
}

func (d *NetworksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state NetworksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.List("networks", "networks")
	if err != nil {
		resp.Diagnostics.AddError("Error listing networks", err.Error())
		return
	}

	for _, item := range items {
		name := fmt.Sprintf("%v", item["nasname"])
		if !state.NameFilter.IsNull() && !state.NameFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(state.NameFilter.ValueString())) {
				continue
			}
		}
		state.Items = append(state.Items, NetworkItemModel{
			ID:        stringVal(item, "id"),
			Name:      types.StringValue(name),
			Region:    stringVal(item, "region"),
			AuthPort:  intVal(item, "auth_port"),
			AcctPort:  intVal(item, "acct_port"),
			PrimaryIP: stringVal(item, "primary_ip"),
			Secret:    stringVal(item, "secret"),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
