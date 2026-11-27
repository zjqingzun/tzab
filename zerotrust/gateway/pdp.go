package main

func checkAccess(subject, resource, action string) (bool, string) {
	identity, exists := getIdentity(subject)

	if !exists {
		return false, "identity not found"
	}

	if identity.Status != "ACTIVE" {
		return false, "identity inactive"
	}

	grant, exists := getGrant(subject, resource, action)

	if !exists {
		return false, "no matching active grant"
	}

	if grant.Status != "ACTIVE" {
		return false, "grant revoked"
	}

	return true, "access granted"
}
