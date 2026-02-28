package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/client"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/datasources"
	"github.com/ironwifi/terraform-provider-ironwifi/internal/resources"
)

var _ provider.Provider = &IronWiFiProvider{}

// IronWiFiProvider implements the IronWiFi Terraform provider.
type IronWiFiProvider struct {
	version string
}

// IronWiFiProviderModel maps provider schema data.
type IronWiFiProviderModel struct {
	APIEndpoint  types.String `tfsdk:"api_endpoint"`
	APIToken     types.String `tfsdk:"api_token"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	CompanyID    types.String `tfsdk:"company_id"`
}

// New returns a provider.Provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IronWiFiProvider{version: version}
	}
}

func (p *IronWiFiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ironwifi"
	resp.Version = p.version
}

func (p *IronWiFiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage IronWiFi resources as infrastructure-as-code.",
		Attributes: map[string]schema.Attribute{
			"api_endpoint": schema.StringAttribute{
				Description: "IronWiFi API endpoint URL. Can also be set with IRONWIFI_API_ENDPOINT env var. Defaults to https://console.ironwifi.com.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "API token for authentication. Can also be set with IRONWIFI_API_TOKEN env var. Recommended over username/password.",
				Optional:    true,
				Sensitive:   true,
			},
			"username": schema.StringAttribute{
				Description: "Username for OAuth2 authentication. Can also be set with IRONWIFI_USERNAME env var.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for OAuth2 authentication. Can also be set with IRONWIFI_PASSWORD env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_id": schema.StringAttribute{
				Description: "OAuth2 client ID. Can also be set with IRONWIFI_CLIENT_ID env var.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "OAuth2 client secret. Can also be set with IRONWIFI_CLIENT_SECRET env var.",
				Optional:    true,
				Sensitive:   true,
			},
			"company_id": schema.StringAttribute{
				Description: "Company/tenant ID for multi-tenant isolation. Can also be set with IRONWIFI_COMPANY_ID env var.",
				Optional:    true,
			},
		},
	}
}

func (p *IronWiFiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config IronWiFiProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve values from config or environment
	apiEndpoint := stringValueOrEnv(config.APIEndpoint, "IRONWIFI_API_ENDPOINT", "https://console.ironwifi.com")
	apiToken := stringValueOrEnv(config.APIToken, "IRONWIFI_API_TOKEN", "")
	username := stringValueOrEnv(config.Username, "IRONWIFI_USERNAME", "")
	password := stringValueOrEnv(config.Password, "IRONWIFI_PASSWORD", "")
	clientID := stringValueOrEnv(config.ClientID, "IRONWIFI_CLIENT_ID", "")
	clientSecret := stringValueOrEnv(config.ClientSecret, "IRONWIFI_CLIENT_SECRET", "")
	companyID := stringValueOrEnv(config.CompanyID, "IRONWIFI_COMPANY_ID", "")

	if companyID == "" {
		resp.Diagnostics.AddError("Missing Company ID", "company_id must be set in provider config or IRONWIFI_COMPANY_ID environment variable")
		return
	}

	if apiToken == "" && username == "" {
		resp.Diagnostics.AddError("Missing Authentication", "Either api_token or username/password must be configured")
		return
	}

	c, err := client.New(&client.Config{
		APIEndpoint:  apiEndpoint,
		APIToken:     apiToken,
		Username:     username,
		Password:     password,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CompanyID:    companyID,
		UserAgent:    "terraform-provider-ironwifi/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Configuration Error", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *IronWiFiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewNetworkResource,
		resources.NewUserResource,
		resources.NewGroupResource,
		resources.NewPolicyResource,
		resources.NewAuthProviderResource,
		resources.NewCaptivePortalResource,
		resources.NewDeviceResource,
		resources.NewCertificateResource,
		resources.NewProfileResource,
		resources.NewConnectorResource,
		resources.NewVoucherResource,
		resources.NewOrgUnitResource,
	}
}

func (p *IronWiFiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewNetworksDataSource,
		datasources.NewUsersDataSource,
		datasources.NewGroupsDataSource,
		datasources.NewPoliciesDataSource,
		datasources.NewDevicesDataSource,
		datasources.NewAuthProvidersDataSource,
	}
}

func stringValueOrEnv(val types.String, envKey, defaultVal string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultVal
}
