package models

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Permission string

// Role represents a named set of permissions
type Role string

// RoleArray is a custom type for handling PostgreSQL text arrays
type RoleArray []Role

// Scan implements the sql.Scanner interface for RoleArray
func (ra *RoleArray) Scan(value interface{}) error {
	if value == nil {
		*ra = RoleArray{}
		return nil
	}

	// Use pq.Array to scan the PostgreSQL array
	var strArray pq.StringArray
	if err := strArray.Scan(value); err != nil {
		return err
	}

	// Convert []string to []Role
	roles := make([]Role, len(strArray))
	for i, s := range strArray {
		roles[i] = Role(s)
	}
	*ra = roles
	return nil
}

// Value implements the driver.Valuer interface for RoleArray
func (ra RoleArray) Value() (driver.Value, error) {
	if len(ra) == 0 {
		return pq.Array([]string{}), nil
	}

	// Convert []Role to []string
	strArray := make([]string, len(ra))
	for i, r := range ra {
		strArray[i] = string(r)
	}
	return pq.Array(strArray), nil
}

// GormDataType tells GORM what database type to use
func (ra RoleArray) GormDataType() string {
	return "text[]"
}

// PermissionArray is a custom type for handling PostgreSQL text arrays
type PermissionArray []Permission

// Scan implements the sql.Scanner interface for PermissionArray
func (pa *PermissionArray) Scan(value interface{}) error {
	if value == nil {
		*pa = PermissionArray{}
		return nil
	}

	// Use pq.Array to scan the PostgreSQL array
	var strArray pq.StringArray
	if err := strArray.Scan(value); err != nil {
		return err
	}

	// Convert []string to []Permission
	permissions := make([]Permission, len(strArray))
	for i, s := range strArray {
		permissions[i] = Permission(s)
	}
	*pa = permissions
	return nil
}

// Value implements the driver.Valuer interface for PermissionArray
func (pa PermissionArray) Value() (driver.Value, error) {
	if len(pa) == 0 {
		return pq.Array([]string{}), nil
	}

	// Convert []Permission to []string
	strArray := make([]string, len(pa))
	for i, p := range pa {
		strArray[i] = string(p)
	}
	return pq.Array(strArray), nil
}

// GormDataType tells GORM what database type to use
func (pa PermissionArray) GormDataType() string {
	return "text[]"
}

// Define standard roles
const (
	RoleAdmin   Role = "admin"
	RoleShipper Role = "shipper"
	RoleCarrier Role = "carrier"
)

const (
	// Shipment Permissions
	CreateShipment Permission = "create_shipment"
	ViewShipment   Permission = "view_shipment"
	EditShipment   Permission = "edit_shipment"
	DeleteShipment Permission = "delete_shipment"

	// Rate Permissions
	ManageRates Permission = "manage_rates"
	ViewRates   Permission = "view_rates"
	AddRoutes   Permission = "add_routes"
	ViewRoutes  Permission = "view_routes"

	// User Management
	ManageUsers Permission = "manage_users"
	ViewUsers   Permission = "view_users"

	// Financial Permissions
	ViewFinancials Permission = "view_financials"
	ManagePayments Permission = "manage_payments"

	// Admin Permissions
	SystemAdmin    Permission = "system_admin"
	ManageSettings Permission = "manage_settings"
	ViewSettings   Permission = "view_settings"
)

// DefaultRolePermissions maps roles to their default permissions
var DefaultRolePermissions = map[Role][]Permission{
	RoleAdmin: {
		SystemAdmin, ManageSettings, ViewSettings,
		ManageUsers, ViewUsers,
		CreateShipment, ViewShipment, EditShipment, DeleteShipment,
		ManageRates, ViewRates, AddRoutes, ViewRoutes,
		ViewFinancials, ManagePayments,
	},
	RoleShipper: {
		CreateShipment, ViewShipment, EditShipment,
		ViewRates, ViewRoutes,
		ViewFinancials,
	},
	RoleCarrier: {
		ViewShipment,
		ManageRates, ViewRates, AddRoutes, ViewRoutes,
		ViewFinancials,
	},
}

// HasPermission checks if a role has a specific permission
func (r Role) HasPermission(permission Permission) bool {
	permissions, exists := DefaultRolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// default vales for id
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	// Only generate a new UUID if one hasn't been set
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}

type User struct {
	BaseModel
	Username     string          `json:"username"`
	Email        string          `json:"email" gorm:"uniqueIndex"`
	Password     string          `json:"-"` // Never expose the password in JSON responses
	CompanyID    uuid.UUID       `json:"company_id"`
	Company      *Company        `json:"company" gorm:"foreignKey:CompanyID"`
	Roles        RoleArray       `json:"roles" gorm:"type:text[]"`        // Using custom RoleArray type
	Permissions  PermissionArray `json:"permissions" gorm:"type:text[]"`  // Using custom PermissionArray type
	ProfileImage string          `json:"profile_image,omitempty"`
	Active       bool            `json:"active" gorm:"default:true"`
	LastLogin    *time.Time      `json:"last_login,omitempty"`
}

func (u *User) HasPermission(permission Permission) bool {
	// First check custom permissions assigned directly to the user
	for _, p := range []Permission(u.Permissions) {
		if p == permission {
			return true
		}
	}

	// Then check role-based permissions
	for _, role := range []Role(u.Roles) {
		if role.HasPermission(permission) {
			return true
		}
	}

	return false
}

// Company represents a company in the system
type Company struct {
	BaseModel
	Name           string  `json:"name"`
	Email          string  `json:"email" gorm:"uniqueIndex"`
	Phone          string  `json:"phone,omitempty"`
	Address        string  `json:"address,omitempty"`
	City           string  `json:"city,omitempty"`
	State          string  `json:"state,omitempty"`
	ZipCode        string  `json:"zip_code,omitempty"`
	Country        string  `json:"country,omitempty"`
	LogoURL        string  `json:"logo_url,omitempty"`
	Website        string  `json:"website,omitempty"`
	TaxID          string  `json:"tax_id,omitempty"`
	CompanyType    string  `json:"company_type"` // "shipper", "carrier", "both"
	Active         bool    `json:"active" gorm:"default:true"`
	VerificationID string  `json:"verification_id,omitempty"`
	Verified       bool    `json:"verified" gorm:"default:false"`
	Users          *[]User `json:"users,omitempty" gorm:"foreignKey:CompanyID"`
}

// Contact represents contact form submissions from the website
type Contact struct {
	BaseModel
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Company string `json:"company,omitempty"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Status  string `json:"status" gorm:"default:'new'"` // "new", "read", "responded", "closed"
}

// MailingList represents email subscribers for marketing
type MailingList struct {
	BaseModel
	Email   string `json:"email" gorm:"uniqueIndex"`
	Name    string `json:"name,omitempty"`
	Active  bool   `json:"active" gorm:"default:true"`
	Source  string `json:"source,omitempty"` // "contact_form", "newsletter_signup", etc.
}