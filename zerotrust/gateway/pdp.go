package main

func checkAccess(subject, resource, action string) (bool, string) {
	return checkAccessOnChain(subject, resource, action)
}
