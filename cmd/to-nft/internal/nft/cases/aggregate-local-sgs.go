package cases

import (
	"context"
	"net"
	"sort"
	"sync"

	sgAPI "github.com/H-BF/protos/pkg/api/sgroups"
	conv "github.com/H-BF/sgroups/internal/api/sgroups"
	"github.com/H-BF/sgroups/internal/dict"
	model "github.com/H-BF/sgroups/internal/models/sgroups"

	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/pkg/slice"
	"github.com/c-robinson/iplib"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// Network is type alias
	Network = model.Network

	//SgName is a type alias
	SgName = string

	//SGClient is a type alias
	SGClient = sgAPI.SecGroupServiceClient

	// SG Secutity Group
	SG struct {
		model.SecurityGroup
		LocalIPsV4 iplib.ByIP
		LocalIPsV6 iplib.ByIP
	}

	// SGs Security Groups dictionary indexed by its names
	SGs struct {
		dict.HDict[SgName, *SG]
	}

	// SGsNetworks Secuurity Group Networks dictionary indexed by Name from SG
	SGsNetworks struct {
		dict.HDict[string, []model.Network]
	}
)

// IsLocal gives true if GS contains any IP from local host
func (loc *SG) IsLocal() bool {
	return loc.LocalIPsV4.Len()+loc.LocalIPsV6.Len() > 0
}

// LoadFromNames load SG from its names
func (loc *SGs) LoadFromNames(ctx context.Context, client SGClient, names []string) (err error) {
	const api = "SG(s)/LoadFromNames"

	defer func() {
		err = errors.WithMessage(err, api)
	}()
	if len(names) == 0 {
		return nil
	}
	req := sgAPI.ListSecurityGroupsReq{SgNames: names}
	var resp *sgAPI.ListSecurityGroupsResp
	if resp, err = client.ListSecurityGroups(ctx, &req); err != nil {
		return err
	}
	for _, g := range resp.GetGroups() {
		var sg SG
		if sg.SecurityGroup, err = conv.Proto2ModelSG(g); err != nil {
			return err
		}
		_ = loc.Insert(sg.Name, &sg)
	}
	return nil
}

// LoadFromIPs it loads Local SGs by IPs
func (loc *SGs) LoadFromIPs(ctx context.Context, client SGClient, localIPs []net.IP) error {
	const api = "SG(s)/load-from-IP(s)"

	if len(localIPs) == 0 {
		return nil
	}
	var mx sync.Mutex
	job := func(i int) error {
		srcIP := localIPs[i]
		req := &sgAPI.GetSecGroupForAddressReq{
			Address: srcIP.String(),
		}
		resp, err := client.GetSecGroupForAddress(ctx, req)
		if err != nil {
			if status.Code(errors.Cause(err)) == codes.NotFound {
				return nil
			}
			return err
		}
		var sg model.SecurityGroup
		if sg, err = conv.Proto2ModelSG(resp); err != nil {
			return err
		}
		mx.Lock()
		defer mx.Unlock()
		it := loc.At(sg.Name)
		if it == nil {
			it = &SG{SecurityGroup: sg}
			loc.Put(sg.Name, it)
		}
		switch len(srcIP) {
		case net.IPv4len:
			it.LocalIPsV4 = append(it.LocalIPsV4, srcIP)
		case net.IPv6len:
			it.LocalIPsV6 = append(it.LocalIPsV6, srcIP)
		}
		return nil
	}
	if err := parallel.ExecAbstract(len(localIPs), 8, job); err != nil { //nolint:gomnd
		return errors.WithMessage(err, api)
	}
	loc.Iterate(func(_ string, it *SG) bool {
		for _, ips := range []*iplib.ByIP{&it.LocalIPsV4, &it.LocalIPsV6} {
			sort.Sort(*ips)
			_ = slice.DedupSlice(ips, func(i, j int) bool {
				l, r := (*ips)[i], (*ips)[j]
				return l.Equal(r)
			})
		}
		return true
	})
	return nil
}

// Names get local SG(s) names
func (loc SGs) Names() []SgName {
	return loc.Keys()
}

// IsEq -
func (loc *SGs) IsEq(other SGs) bool {
	return loc.Eq(&other.HDict, func(lSg, rSg *SG) bool {
		return lSg.IsEq(rSg.SecurityGroup)
	})
}

// IsEq -
func (sgsNws *SGsNetworks) IsEq(other SGsNetworks) bool {
	eq := sgsNws.Len() == other.Len()
	if eq {
		var x, y dict.Dict[string, *model.Network]
		eq = sgsNws.Eq(&other, func(vL, vR []model.Network) bool {
			if len(vL) == len(vR) {
				x.Clear()
				y.Clear()
				for i := range vL {
					_ = x.Insert(vL[i].Name, &vL[i])
				}
				for i := range vR {
					_ = y.Insert(vR[i].Name, &vR[i])
				}
				return x.Eq(y, func(nwL, nwR *model.Network) bool {
					return nwL.IsEq(*nwR)
				})
			}
			return false
		})
	}
	return eq
}

// LoadFromSGNames -
func (sgsNws *SGsNetworks) LoadFromSGNames(ctx context.Context, client SGClient, sgNames []string) error {
	const api = "SG(s)/Networks/Load"

	sgsNws.Clear()
	var mx sync.Mutex
	err := parallel.ExecAbstract(len(sgNames), 8, func(i int) error { //nolint:gomnd
		req := sgAPI.GetSgSubnetsReq{SgName: sgNames[i]}
		resp, e := client.GetSgSubnets(ctx, &req)
		if e != nil {
			if status.Code(errors.Cause(e)) == codes.NotFound {
				return nil
			}
			return e
		}
		protoNws := resp.GetNetworks()
		nws := make([]model.Network, 0, len(protoNws))
		for _, nw := range protoNws {
			var m model.Network
			if m, e = conv.Proto2ModelNetwork(nw); e != nil {
				return e
			}
			nws = append(nws, m)
		}
		mx.Lock()
		defer mx.Unlock()
		sgsNws.Put(sgNames[i], nws)
		return nil
	})
	return errors.WithMessage(err, api)
}
