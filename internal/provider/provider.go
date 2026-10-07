package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xtruong/terraform-provider-superset/internal/client"
)

var (
	_ provider.Provider = &SupersetProvider{}
)

// SupersetProvider defines the provider implementation.
type SupersetProvider struct {
	version string
}

// SupersetProviderModel maps provider schema data.
type SupersetProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// New returns a new provider implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SupersetProvider{
			version: version,
		}
	}
}

func (p *SupersetProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "superset"
	resp.Version = p.version
}

func (p *SupersetProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider to manage Apache Superset users, roles, and permissions (RBAC).",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "The root URL of the Apache Superset instance (e.g. http://localhost:8088). Can also be set via SUPERSET_BASE_URL environment variable.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for authenticating against the Superset security API. Can also be set via SUPERSET_USERNAME environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for authenticating against the Superset security API. Can also be set via SUPERSET_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Disable TLS verification (useful for self-signed certificates or local testing). Defaults to false.",
				Optional:    true,
			},
		},
	}
}

func (p *SupersetProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SupersetProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := os.Getenv("SUPERSET_BASE_URL")
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	username := os.Getenv("SUPERSET_USERNAME")
	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		username = config.Username.ValueString()
	}

	password := os.Getenv("SUPERSET_PASSWORD")
	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		password = config.Password.ValueString()
	}

	insecure := false
	if !config.Insecure.IsNull() && !config.Insecure.IsUnknown() {
		insecure = config.Insecure.ValueBool()
	} else if os.Getenv("SUPERSET_INSECURE") == "true" {
		insecure = true
	}

	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing Superset Base URL",
			"The provider cannot configure the Superset client because base_url was not provided in the configuration or SUPERSET_BASE_URL environment variable.",
		)
		return
	}

	c, err := client.NewClient(baseURL, username, password, insecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Superset API client",
			"An unexpected error occurred while creating the client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *SupersetProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewRoleResource,
		NewRowLevelSecurityResource,
	}
}

func (p *SupersetProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewRoleDataSource,
		NewPermissionDataSource,
	}
}

