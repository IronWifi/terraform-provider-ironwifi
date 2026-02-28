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

var _ datasource.DataSource = &UsersDataSource{}

// UsersDataSource lists IronWiFi users.
type UsersDataSource struct {
	client *client.Client
}

// UsersDataSourceModel is the schema model for the users data source.
type UsersDataSourceModel struct {
	UsernameFilter types.String    `tfsdk:"username_filter"`
	Items          []UserItemModel `tfsdk:"items"`
}

// UserItemModel represents a single user in the data source output.
type UserItemModel struct {
	ID           types.String `tfsdk:"id"`
	Username     types.String `tfsdk:"username"`
	Email        types.String `tfsdk:"email"`
	Firstname    types.String `tfsdk:"firstname"`
	Lastname     types.String `tfsdk:"lastname"`
	UserType     types.String `tfsdk:"user_type"`
	CreationDate types.String `tfsdk:"creationdate"`
}

// NewUsersDataSource returns a new users data source.
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists IronWiFi users.",
		Attributes: map[string]schema.Attribute{
			"username_filter": schema.StringAttribute{
				Description: "Optional substring filter applied to usernames.",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The list of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "User ID."},
						"username":     schema.StringAttribute{Computed: true, Description: "Username."},
						"email":        schema.StringAttribute{Computed: true, Description: "Email address."},
						"firstname":    schema.StringAttribute{Computed: true, Description: "First name."},
						"lastname":     schema.StringAttribute{Computed: true, Description: "Last name."},
						"user_type":    schema.StringAttribute{Computed: true, Description: "User type."},
						"creationdate": schema.StringAttribute{Computed: true, Description: "Creation date."},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, err := d.client.List("users", "users")
	if err != nil {
		resp.Diagnostics.AddError("Error listing users", err.Error())
		return
	}

	for _, item := range items {
		username := fmt.Sprintf("%v", item["username"])
		if !state.UsernameFilter.IsNull() && !state.UsernameFilter.IsUnknown() {
			if !strings.Contains(strings.ToLower(username), strings.ToLower(state.UsernameFilter.ValueString())) {
				continue
			}
		}
		state.Items = append(state.Items, UserItemModel{
			ID:           stringVal(item, "id"),
			Username:     types.StringValue(username),
			Email:        stringVal(item, "email"),
			Firstname:    stringVal(item, "firstname"),
			Lastname:     stringVal(item, "lastname"),
			UserType:     stringVal(item, "user_type"),
			CreationDate: stringVal(item, "creationdate"),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
