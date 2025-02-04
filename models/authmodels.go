package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission string

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
	Username    string       `json:"username"`
	Email       string       `json:"email"`
	Password    string       `json:"password"`
	CompanyID   uuid.UUID    `json:"company_id"`
	Company     *Company     `json:"company" gorm:"foreignKey:CompanyID"`
	Permissions []Permission `gorm:"type:text[]"` // Using PostgreSQL text array
}

type Company struct {
	BaseModel
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Users    *[]User `json:"users" gorm:"foreignKey:CompanyID"`
}
