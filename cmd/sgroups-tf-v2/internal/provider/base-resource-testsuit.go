package provider

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	sgAPI "github.com/H-BF/sgroups/internal/api/sgroups"
	client "github.com/H-BF/sgroups/internal/grpc-client"
	registry "github.com/H-BF/sgroups/internal/registry/sgroups"
	"google.golang.org/grpc"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/suite"
)

type baseResourceTests struct {
	suite.Suite

	ctx               context.Context
	ctxCancel         func()
	grpcServer        *grpc.Server
	sgClient          sgAPI.ClosableClient
	providerConfig    string
	providerFactories map[string]func() (tfprotov6.ProviderServer, error)
}

func (sui *baseResourceTests) SetupSuite() {
	sui.ctx = context.Background()
	if dl, ok := sui.T().Deadline(); ok {
		sui.ctx, sui.ctxCancel = context.WithDeadline(sui.ctx, dl)
	}

	addr := path.Join("/tmp", fmt.Sprintf("test-%v-%v.socket", os.Getpid(), time.Now().Nanosecond()))
	var opts []grpc.ServerOption
	sui.grpcServer = grpc.NewServer(opts...)
	errc := make(chan struct{}, 1)
	go runServerDetached(sui.ctx, sui.grpcServer, addr, errc)
	fmt.Printf("server started on %s\n", addr)

	address := fmt.Sprintf("unix://%s", addr)
	os.Setenv("SGROUPS_ADDRESS", address)
	sui.providerConfig = `
		provider "sgroups" {
			address = ` + address + `
		}
		`
	dialDuration := lookupEnvWithDefault("SGROUPS_DIAL_DURATION", "10s")
	connDuration, err := time.ParseDuration(dialDuration)
	sui.Require().Nil(err)

	builder := client.FromAddress(address).
		WithDialDuration(connDuration)

	sui.sgClient, err = sgAPI.NewClosableClient(sui.ctx, builder)
	sui.Require().NoError(err)

	sui.providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sgroups": providerserver.NewProtocol6WithError(Factory("test")()),
	}
}

func (sui *baseResourceTests) TearDownSuite() {
	err := sui.sgClient.CloseConn()
	sui.Require().NoError(err)

	sui.grpcServer.Stop()
	fmt.Println("server stopped")

	if sui.ctxCancel != nil {
		sui.ctxCancel()
	}
}

func slice2string[T fmt.Stringer](args ...T) string {
	data := bytes.NewBuffer(nil)
	for i, o := range args {
		if i > 0 {
			_, _ = data.WriteString(";  ")
		}
		_, _ = fmt.Fprintf(data, "%s", o)
	}
	return strings.ReplaceAll(data.String(), `"`, "'")
}

func runServerDetached(ctx context.Context, grpcServer *grpc.Server, addr string, errc chan struct{}) {
	lis, err := net.Listen("unix", addr)
	if err != nil {
		fmt.Printf("cant listen on %s: %s\n", addr, err.Error())
		errc <- struct{}{}
		return
	}

	m, err := registry.NewMemDB(registry.TblSecGroups,
		registry.TblSecRules, registry.TblNetworks,
		registry.TblSyncStatus, registry.TblFqdnRules,
		registry.IntegrityChecker4SG(),
		registry.IntegrityChecker4SGRules(),
		registry.IntegrityChecker4FqdnRules(),
		registry.IntegrityChecker4Networks())
	if err != nil {
		fmt.Printf("cant create mem db : %s\n", err.Error())
		errc <- struct{}{}
		return
	}
	service := sgAPI.NewSGroupsService(ctx, registry.NewRegistryFromMemDB(m))

	service.RegisterGRPC(ctx, grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println("server stopped due" + err.Error())
		errc <- struct{}{}
	}
}
