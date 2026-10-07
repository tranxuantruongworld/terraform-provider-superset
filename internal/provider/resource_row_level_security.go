package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/xtruong/terraform-provider-superset/internal/client"
)

var (
	_ resource.Resource                = &rlsResource{}
	_ resource.ResourceWithConfigure   = &rlsResource{}
	_ resource.ResourceWithImportState = &rlsResource{}
)

func NewRowLevelSecurityResource() resource.Resource {
	return &rlsResource{}
}

type rlsResource struct {
	client *client.Client
}

type rlsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	FilterType  types.String `tfsdk:"filter_type"`
	Clause      types.String `tfsdk:"clause"`
	Tables      types.List   `tfsdk:"tables"`
	Roles       types.List   `tfsdk:"roles"`
}

func (r *rlsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_row_level_security"
}

func (r *rlsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Row-Level Security (RLS) filter rule within Apache Superset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the RLS rule as a string.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the RLS rule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description explaining what this RLS rule filters.",
				Optional:    true,
			},
			"filter_type": schema.StringAttribute{
				Description: "The filter type: 'Regular' (applies to specified roles) or 'Base' (applies to all except specified roles). Defaults to 'Regular'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Regular"),
			},
			"clause": schema.StringAttribute{
				Description: "The SQL filter clause injected into queries (e.g., \"country = 'VN'\" or \"department_id = 12\").",
				Required:    true,
			},
			"tables": schema.ListAttribute{
				Description: "List of dataset/table IDs to which this RLS rule applies.",
				ElementType: types.Int64Type,
				Required:    true,
			},
			"roles": schema.ListAttribute{
				Description: "List of role IDs to associate with this RLS rule.",
				ElementType: types.Int64Type,
				Optional:    true,
			},
		},
	}
}

func (r *rlsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *rlsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rlsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tableIDs []int64
	diags = plan.Tables.ElementsAs(ctx, &tableIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var roleIDs []int64
	if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
		diags = plan.Roles.ElementsAs(ctx, &roleIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	desc := ""
	if !plan.Description.IsNull() {
		desc = plan.Description.ValueString()
	}

	createReq := &client.CreateRLSRequest{
		Name:        plan.Name.ValueString(),
		Description: desc,
		FilterType:  plan.FilterType.ValueString(),
		Clause:      plan.Clause.ValueString(),
		Tables:      tableIDs,
		Roles:       roleIDs,
	}

	rls, err := r.client.CreateRLS(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Row-Level Security Rule",
			"Could not create RLS rule, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(rls.ID, 10))
	plan.Name = types.StringValue(rls.Name)
	plan.FilterType = types.StringValue(rls.FilterType)
	plan.Clause = types.StringValue(rls.Clause)
	if rls.Description != "" {
		plan.Description = types.StringValue(rls.Description)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *rlsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rlsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid RLS Rule ID", "Could not parse RLS ID: "+err.Error())
		return
	}

	rls, err := r.client.GetRLS(ctx, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Row-Level Security Rule",
			fmt.Sprintf("Could not read RLS rule %d: %s", id, err.Error()),
		)
		return
	}

	state.Name = types.StringValue(rls.Name)
	state.FilterType = types.StringValue(rls.FilterType)
	state.Clause = types.StringValue(rls.Clause)
	if rls.Description != "" {
		state.Description = types.StringValue(rls.Description)
	}

	var tableIDs []int64
	for _, t := range rls.Tables {
		tableIDs = append(tableIDs, t.ID)
	}
	tablesVal, diags := types.ListValueFrom(ctx, types.Int64Type, tableIDs)
	resp.Diagnostics.Append(diags...)
	state.Tables = tablesVal

	var roleIDs []int64
	for _, role := range rls.Roles {
		roleIDs = append(roleIDs, role.ID)
	}
	rolesVal, diags := types.ListValueFrom(ctx, types.Int64Type, roleIDs)
	resp.Diagnostics.Append(diags...)
	state.Roles = rolesVal

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *rlsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rlsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid RLS Rule ID", "Could not parse RLS ID: "+err.Error())
		return
	}

	var tableIDs []int64
	diags = plan.Tables.ElementsAs(ctx, &tableIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var roleIDs []int64
	if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
		diags = plan.Roles.ElementsAs(ctx, &roleIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	name := plan.Name.ValueString()
	filterType := plan.FilterType.ValueString()
	clause := plan.Clause.ValueString()

	updateReq := &client.UpdateRLSRequest{
		Name:       &name,
		FilterType: &filterType,
		Clause:     &clause,
		Tables:     tableIDs,
		Roles:      roleIDs,
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	rls, err := r.client.UpdateRLS(ctx, id, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Row-Level Security Rule",
			fmt.Sprintf("Could not update RLS rule %d: %s", id, err.Error()),
		)
		return
	}

	plan.Name = types.StringValue(rls.Name)
	plan.FilterType = types.StringValue(rls.FilterType)
	plan.Clause = types.StringValue(rls.Clause)
	if rls.Description != "" {
		plan.Description = types.StringValue(rls.Description)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *rlsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rlsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid RLS Rule ID", "Could not parse RLS ID: "+err.Error())
		return
	}

	if err := r.client.DeleteRLS(ctx, id); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Row-Level Security Rule",
			fmt.Sprintf("Could not delete RLS rule %d: %s", id, err.Error()),
		)
		return
	}
}

func (r *rlsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
