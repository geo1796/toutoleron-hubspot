package crm

const (
	UserInternalName = "users"
	UserObjectTypeID = "0-115"

	UserPropertyInternalUserID = "hs_internal_user_id"
	UserPropertyOwnerID        = "hubspot_owner_id"
)

type User struct {
	Object
}

func (u *User) OwnerID() string {
	return u.GetProperty(UserPropertyOwnerID)
}
