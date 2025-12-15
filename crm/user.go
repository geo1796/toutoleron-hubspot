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

func NewUser() *User {
	return &User{
		Object: NewBaseObject(
			UserInternalName,
			UserObjectTypeID,
		),
	}
}

func (u *User) OwnerID() string {
	return u.GetProperty(UserPropertyOwnerID)
}
