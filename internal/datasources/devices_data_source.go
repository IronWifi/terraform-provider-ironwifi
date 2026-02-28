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

var _ datasource.DataSource = &DevicesDataSource{}

// DevicesDataSource lists IronWiFi devices.
type DevicesDataSource struct {
	client *client.Client
}

// DevicesDataSourceModel is the schema model for the devices data source.
type DevicesDataSourceModel struct {
	NameFilter types.String      `tfsdk:"name_filter"`
	Items      []DeviceItemModel `tfsdk:"items"`
}

// DeviceItemModel represents a single device in the data source output.
type DeviceItemModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Email        types.String `tfsdk:"email"`
	Firstname    types.String `tfsdk:"firstname"`
	Lastname     types.String `tfsdk:"lastname"`
	CreationDate types.String `tfsdk:"creationdate"`
}

// NewDevicesDataSource returns a new devices data source.
func NewDevicesDataSource() datasource.DataSource {
	return &DevicesDataSource{}
}

func (d *DevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *DevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists IronWiFi devices.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to device names.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of devices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "Device ID."},
						"name":         schema.StringAttribute{Computed: true, Description: "Device name (username)."},
						"email":        schema.StringAttribute{Computed: true, Description: "Email address."},
						"firstname":    schema.StringAttribute{Computed: true, Description: "First name."},
						"lastname":     schema.StringAttribute{Computed: true, Description: "Last name."},
						"creationdate": schema.StringAttribute{Computed: true, Description: "Creation date."},
					},
				},
			},
		},
	}
}

func (d *DevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.List("devices", "devices")
	if err != nil {
		resp.Diagnostics.AddError("Error listing devices", err.Error())
		return
	}

	for _, item := range items {
		name := fmt.Sprintf("%v", item["username"])
		if !state.NameFilter.IsNull() && !state.NameFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(state.NameFilter.ValueString())) {
				continue
			}
		}
		state.Items = append(state.Items, DeviceItemModel{
			ID:           stringVal(item, "id"),
			Name:         types.StringValue(name),
			Email:        stringVal(item, "email"),
			Firstname:    stringVal(item, "firstname"),
			Lastname:     stringVal(item, "lastname"),
			CreationDate: stringVal(item, "creationdate"),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
