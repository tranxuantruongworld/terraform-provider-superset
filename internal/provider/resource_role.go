package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xtruong/terraform-provider-superset/internal/client"
)

var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *client.Client
}

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.List   `tfsdk:"permissions"`
}

func (r *roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a role and its permission view menus (PVM) within Apache Superset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the role as a string.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the role (e.g., 'Finance_Analyst', 'Data_Engineer').",
				Required:    true,
			},
			"permissions": schema.ListAttribute{
				Description: "List of permission view menu (PVM) IDs assigned to this role.",
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permIDs []int64
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		diags = plan.Permissions.ElementsAs(ctx, &permIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := &client.CreateRoleRequest{
		Name:        plan.Name.ValueString(),
		Permissions: permIDs,
	}

	role, err := r.client.CreateRole(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Superset Role",
			"Could not create role, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(role.ID, 10))
	plan.Name = types.StringValue(role.Name)

	var currentPermIDs []int64
	for _, p := range role.Permissions {
		currentPermIDs = append(currentPermIDs, p.ID)
	}
	permsVal, diags := types.ListValueFrom(ctx, types.Int64Type, currentPermIDs)
	resp.Diagnostics.Append(diags...)
	plan.Permissions = permsVal

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Role ID", "Could not parse role ID: "+err.Error())
		return
	}

	role, err := r.client.GetRole(ctx, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Superset Role",
			fmt.Sprintf("Could not read role with ID %d: %s", id, err.Error()),
		)
		return
	}

	state.Name = types.StringValue(role.Name)

	var permIDs []int64
	for _, p := range role.Permissions {
		permIDs = append(permIDs, p.ID)
	}
	permsVal, diags := types.ListValueFrom(ctx, types.Int64Type, permIDs)
	resp.Diagnostics.Append(diags...)
	state.Permissions = permsVal

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Role ID", "Could not parse role ID: "+err.Error())
		return
	}

	var permIDs []int64
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		diags = plan.Permissions.ElementsAs(ctx, &permIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateReq := &client.UpdateRoleRequest{
		Name:        plan.Name.ValueString(),
		Permissions: permIDs,
	}

	role, err := r.client.UpdateRole(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Superset Role",
			fmt.Sprintf("Could not update role %d: %s", id, err.Error()),
		)
		return
	}

	plan.Name = types.StringValue(role.Name)

	var currentPermIDs []int64
	for _, p := range role.Permissions {
		currentPermIDs = append(currentPermIDs, p.ID)
	}
	permsVal, diags := types.ListValueFrom(ctx, types.Int64Type, currentPermIDs)
	resp.Diagnostics.Append(diags...)
	plan.Permissions = permsVal

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Role ID", "Could not parse role ID: "+err.Error())
		return
	}

	if err := r.client.DeleteRole(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Superset Role",
			fmt.Sprintf("Could not delete role %d: %s", id, err.Error()),
		)
		return
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
