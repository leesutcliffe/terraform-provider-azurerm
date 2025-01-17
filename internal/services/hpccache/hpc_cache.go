package hpccache

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storagecache/mgmt/2021-09-01/storagecache"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func CacheGetAccessPolicyRuleByScope(policyRules []storagecache.NfsAccessRule, scope storagecache.NfsAccessRuleScope) (storagecache.NfsAccessRule, bool) {
	for _, rule := range policyRules {
		if rule.Scope == scope {
			return rule, true
		}
	}

	return storagecache.NfsAccessRule{}, false
}

func CacheGetAccessPolicyByName(policies []storagecache.NfsAccessPolicy, name string) *storagecache.NfsAccessPolicy {
	for _, policy := range policies {
		if policy.Name != nil && *policy.Name == name {
			return &policy
		}
	}
	return nil
}

func CacheDeleteAccessPolicyByName(policies []storagecache.NfsAccessPolicy, name string) []storagecache.NfsAccessPolicy {
	var newPolicies []storagecache.NfsAccessPolicy
	for _, policy := range policies {
		if policy.Name != nil && *policy.Name != name {
			newPolicies = append(newPolicies, policy)
		}
	}
	return newPolicies
}

func CacheInsertOrUpdateAccessPolicy(policies []storagecache.NfsAccessPolicy, policy storagecache.NfsAccessPolicy) ([]storagecache.NfsAccessPolicy, error) {
	if policy.Name == nil {
		return nil, fmt.Errorf("the name of the HPC Cache access policy is nil")
	}
	var newPolicies []storagecache.NfsAccessPolicy

	isNew := true
	for _, existPolicy := range policies {
		if existPolicy.Name != nil && *existPolicy.Name == *policy.Name {
			newPolicies = append(newPolicies, policy)
			isNew = false
			continue
		}
		newPolicies = append(newPolicies, existPolicy)
	}

	if !isNew {
		return newPolicies, nil
	}

	return append(newPolicies, policy), nil
}

func resourceHPCCacheWaitForCreating(ctx context.Context, client *storagecache.CachesClient, resourceGroup, name string, d *pluginsdk.ResourceData) (storagecache.Cache, error) {
	state := &pluginsdk.StateChangeConf{
		MinTimeout: 30 * time.Second,
		Delay:      10 * time.Second,
		Pending:    []string{string(storagecache.ProvisioningStateTypeCreating)},
		Target:     []string{string(storagecache.ProvisioningStateTypeSucceeded)},
		Refresh:    resourceHPCCacheRefresh(ctx, client, resourceGroup, name),
		Timeout:    d.Timeout(pluginsdk.TimeoutCreate),
	}

	resp, err := state.WaitForStateContext(ctx)
	if err != nil {
		return resp.(storagecache.Cache), fmt.Errorf("waiting for the HPC Cache %q to be missing (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return resp.(storagecache.Cache), nil
}

func resourceHPCCacheRefresh(ctx context.Context, client *storagecache.CachesClient, resourceGroup, name string) pluginsdk.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return resp, "NotFound", nil
			}

			return resp, "Error", fmt.Errorf("making Read request on HPC Cache %q (Resource Group %q): %+v", name, resourceGroup, err)
		}

		return resp, string(resp.ProvisioningState), nil
	}
}
