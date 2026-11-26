package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		IdentityMap: []Identity{}, ResourceMap: []Resource{}, GrantMap: []Grant{}, PolicyMap: []Policy{}, AuditRecordList: []AuditRecord{}}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	identityIndexMap := make(map[string]struct{})

	for _, elem := range gs.IdentityMap {
		index := fmt.Sprint(elem.Did)
		if _, ok := identityIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for identity")
		}
		identityIndexMap[index] = struct{}{}
	}
	resourceIndexMap := make(map[string]struct{})

	for _, elem := range gs.ResourceMap {
		index := fmt.Sprint(elem.ResourceId)
		if _, ok := resourceIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for resource")
		}
		resourceIndexMap[index] = struct{}{}
	}
	grantIndexMap := make(map[string]struct{})

	for _, elem := range gs.GrantMap {
		index := fmt.Sprint(elem.GrantId)
		if _, ok := grantIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for grant")
		}
		grantIndexMap[index] = struct{}{}
	}
	policyIndexMap := make(map[string]struct{})

	for _, elem := range gs.PolicyMap {
		index := fmt.Sprint(elem.PolicyId)
		if _, ok := policyIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for policy")
		}
		policyIndexMap[index] = struct{}{}
	}
	auditRecordIdMap := make(map[uint64]bool)
	auditRecordCount := gs.GetAuditRecordCount()
	for _, elem := range gs.AuditRecordList {
		if _, ok := auditRecordIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for auditRecord")
		}
		if elem.Id >= auditRecordCount {
			return fmt.Errorf("auditRecord id should be lower or equal than the last id")
		}
		auditRecordIdMap[elem.Id] = true
	}

	return gs.Params.Validate()
}
