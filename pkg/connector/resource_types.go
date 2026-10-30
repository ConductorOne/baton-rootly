package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
)

var (
	userResourceType = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
		// Users have no entitlements of their own, but they are the source of the role and
		// on-call role grants: Rootly returns each user's role relationships inline on
		// /v1/users, so the user syncer emits those grants in a single type-scoped pass
		// rather than fanning out one request per user or per role.
		Annotations: annotations.New(
			&v2.SkipEntitlements{},
			&v2.TypeScopedGrants{},
		),
	}
	teamResourceType = &v2.ResourceType{
		Id:          "team",
		DisplayName: "Team",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
	}
	secretResourceType = &v2.ResourceType{
		Id:          "secret",
		DisplayName: "Secret",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_SECRET},
		Annotations: annotations.New(&v2.SkipEntitlementsAndGrants{}),
	}
	scheduleResourceType = &v2.ResourceType{
		Id:          "schedule",
		DisplayName: "Schedule",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
	}
	roleResourceType = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Incident Response Role",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_ROLE},
	}
	onCallRoleResourceType = &v2.ResourceType{
		Id:          "on_call_role",
		DisplayName: "On-Call Role",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_ROLE},
	}
)
