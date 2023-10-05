package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/H-BF/protos/pkg/api/common"
	protos "github.com/H-BF/protos/pkg/api/sgroups"
	"github.com/H-BF/sgroups/cmd/tf-provider-framework/internal/validators"
	sgAPI "github.com/H-BF/sgroups/internal/api/sgroups"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

func NewNetworksResource() resource.Resource {
	return &networksResource{}
}

var _ resource.Resource = &networksResource{}
var _ resource.ResourceWithConfigure = &networksResource{}

type (
	networksResource struct {
		client sgAPI.Client
	}

	networksResourceModel struct {
		Items map[string]networkItem `tfsdk:"items"`
	}

	networkItem struct {
		Cidr types.String `tfsdk:"cidr"`
	}
)

func (nr *networksResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (nr *networksResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Collection of networks",
		Attributes: map[string]schema.Attribute{
			"items": schema.MapNestedAttribute{
				Description: "Mapping from network name to it features (for example CIDR)",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cidr": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								validators.IsCIDR(),
							},
						},
					},
				},
			},
		},
	}
}

func (nr *networksResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(sgAPI.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected sgroups GRPC client, got: %T.", req.ProviderData),
		)

		return
	}

	nr.client = client
}

func (nr *networksResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networksResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into GRPC data model
	syncReq := plan.asSyncReq(protos.SyncReq_Upsert)

	// Send GRPC request
	if _, err := nr.client.Sync(ctx, &syncReq); err != nil {
		resp.Diagnostics.AddError(
			"Error creating networks",
			"Could not create networks: "+err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (nr *networksResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networksResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into GRPC data model
	listReq := state.asReadReq()

	// Create GRPC request
	listResp, err := nr.client.ListNetworks(ctx, &listReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading networks state",
			"Could not perform ListNetworks GRPC call: "+err.Error(),
		)
		return
	}

	// Convert from the GRPC data model to the Terraform data model
	// and refresh any attribute values.
	if srcNetworks := listResp.GetNetworks(); len(srcNetworks) == 0 {
		resp.Diagnostics.AddError(
			"Error reading networks state",
			fmt.Sprintf("Resource doesn't contain at least one of theese networks: %s",
				strings.Join(maps.Keys(state.Items), ", ")))
		return
	}

	if err := state.loadFromProto(listResp); err != nil {
		resp.Diagnostics.AddError(
			"Error reading networks state",
			"Could not convert ListNetworks response into new state: "+err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (nr *networksResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state networksResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	itemsToDelete := map[string]networkItem{}
	for name, networkFeatures := range state.Items {
		if _, ok := plan.Items[name]; !ok {
			// if network is missing in plan state - delete it
			itemsToDelete[name] = networkFeatures
		}
	}

	if len(itemsToDelete) > 0 {
		tempModel := networksResourceModel{
			Items: itemsToDelete,
		}
		delReq := tempModel.asSyncReq(protos.SyncReq_Delete)

		if _, err := nr.client.Sync(ctx, &delReq); err != nil {
			resp.Diagnostics.AddError(
				"Error updating networks",
				"Could not delete networks: "+err.Error())
			return
		}
	}

	createReq := plan.asSyncReq(protos.SyncReq_Upsert)

	if _, err := nr.client.Sync(ctx, &createReq); err != nil {
		resp.Diagnostics.AddError(
			"Error updating networks",
			"Could not create networks: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (nr *networksResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networksResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delReq := state.asSyncReq(protos.SyncReq_Delete)

	if _, err := nr.client.Sync(ctx, &delReq); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting networks",
			"Could not delete networks: "+err.Error())
		return
	}
}

func (model *networksResourceModel) asSyncReq(operation protos.SyncReq_SyncOp) protos.SyncReq {
	var sn protos.SyncNetworks
	for name, netFeatures := range model.Items {
		sn.Networks = append(sn.Networks, &protos.Network{
			Name:    name,
			Network: &common.Networks_NetIP{CIDR: netFeatures.Cidr.ValueString()},
		})
	}
	return protos.SyncReq{
		SyncOp: operation,
		Subject: &protos.SyncReq_Networks{
			Networks: &sn,
		},
	}
}

func (model *networksResourceModel) asReadReq() protos.ListNetworksReq {
	return protos.ListNetworksReq{
		NeteworkNames: maps.Keys(model.Items),
	}
}

func (model *networksResourceModel) loadFromProto(resp *protos.ListNetworksResp) error {
	newItems := make(map[string]networkItem, len(model.Items))
	for _, nw := range resp.GetNetworks() {
		if nw != nil {
			newItems[nw.GetName()] = networkItem{Cidr: types.StringValue(nw.GetNetwork().GetCIDR())}
		}
	}
	return nil
}
