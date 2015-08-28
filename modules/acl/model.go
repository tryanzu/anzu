package acl

type AclRole struct {
	Permissions []string `json:"permissions"`
	Inherits    []string `json:"parents"`
}
