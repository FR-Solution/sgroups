package cases

import (
	"context"

	conv "github.com/H-BF/sgroups/internal/api/sgroups"
	"github.com/H-BF/sgroups/internal/dict"
	model "github.com/H-BF/sgroups/internal/models/sgroups"

	sgAPI "github.com/H-BF/protos/pkg/api/sgroups"
	"github.com/pkg/errors"
)

// SgIeSgRules -
type SgIeSgRules struct {
	SGs   SGs
	Rules dict.HDict[model.SgSgRuleIdentity, *model.SgSgRule]
}

// IsEq -
func (rules *SgIeSgRules) IsEq(other SgIeSgRules) bool {
	eq := rules.SGs.IsEq(other.SGs)
	if eq {
		eq = rules.Rules.Eq(&other.Rules, func(vL, vR *model.SgSgRule) bool {
			return vL.IsEq(*vR)
		})
	}
	return eq
}

// GetRulesForTrafficAndSG -
func (rules *SgIeSgRules) GetRulesForTrafficAndSG(tr model.Traffic, sg string) (ret []*model.SgSgRule) {
	rules.Rules.Iterate(func(k model.SgSgRuleIdentity, v *model.SgSgRule) bool {
		if k.Traffic == tr && k.SgLocal == sg {
			ret = append(ret, v)
		}
		return true
	})
	return ret
}

// Load -
func (rules *SgIeSgRules) Load(ctx context.Context, client SGClient, locals SGs) (err error) {
	const api = "sg-ie-sg-rules/Load"

	defer func() {
		err = errors.WithMessage(err, api)
	}()

	rules.Rules.Clear()
	rules.SGs.Clear()
	localSgNames := locals.Names()
	if len(localSgNames) == 0 {
		return nil
	}
	req := sgAPI.FindSgSgRulesReq{SgLocal: localSgNames}
	var resp *sgAPI.SgSgRulesResp
	if resp, err = client.FindSgSgRules(ctx, &req); err != nil {
		return err
	}
	var nonLocalSgs []string
	for _, protoRule := range resp.GetRules() {
		var rule model.SgSgRule
		if rule, err = conv.Proto2ModelSgSgRule(protoRule); err != nil {
			return err
		}
		_ = rules.Rules.Insert(rule.ID, &rule)
		loc := locals.At(rule.ID.SgLocal)
		_ = rules.SGs.Insert(loc.Name, loc)
		if loc = locals.At(rule.ID.Sg); loc != nil {
			_ = rules.SGs.Insert(loc.Name, loc)
		} else {
			nonLocalSgs = append(nonLocalSgs, rule.ID.Sg)
		}
	}
	return rules.SGs.LoadFromNames(ctx, client, nonLocalSgs)
}
