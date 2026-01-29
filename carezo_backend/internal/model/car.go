package models

import "time"

type Car struct {
	ID 						string `json:"id" db:"id"`

	Model 					string  `json:"model" db:"model"`
	Brand					string  `json:"brand" db:"brand"`
	Year					int		`json:"year"  db:"year"`
	Color   				string	`json:"color" db:"color"`
	LicencePlate			string	`json:"licence_plate" db:"license_plate"`


	EngineOutput     	   *string `json:"engine_output,omitempty" db:"engine_output"` 
	Transmission			string	 `json:"transmission" db:"transmission"`
	FuelType				string   `json:"fuel_type" db:"fuel_type"`
	SeatingCapacity 		int      `json:"seating_capacity" db:"seating_capacity"`
	MaximumSpeed   		   *int      `json:"maximum_speed,omitempty" db:"maximum_speed"`
	Mileage         		int     `json:"mileage" db:"mileage"`                           // total kilometers driven

	DriverName   			*string `json:"driver_name,omitempty" db:"driver_name"`     
	DriverNumber 			*string `json:"driver_number,omitempty" db:"driver_number"` 
	DriverMiles  			*int    `json:"driver_miles,omitempty" db:"driver_miles"`   

	HourlyRate   			JSONB    `json:"hourly_rate" db:"hourly_rate"` 
	CautionFee   			float64  `json:"caution_fee" db:"caution_fee"`

	Features  			JSONB    `json:"features" db:"features"`

	Images       			JSONB     `json:"images" db:"images"`

	IsAvailable  			bool      `json:"is_available" db:"is_available"`
	Status       			string    `json:"status" db:"status"`
	CurrentLocation 	   *string    `json:"current_location, omitempty" db:"current_location"`

	CreatedAt 				time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt 				time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt 			   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type JSONB map[string]interface{}

type HourlyRate struct {
	Standard  float64  `json:"standard"`
	Weekend   float64  `json:"weekend"`
	Holiday   float64  `json:"holiday"`
}


type CreateCarRequest struct {
	Model           string   `json:"model" binding:"required"`
	Brand           string   `json:"brand" binding:"required"`
	Year            int      `json:"year" binding:"required,min=1900"`
	Color           string   `json:"color" binding:"required"`
	LicencePlate    string   `json:"license_plate" binding:"required"`
	EngineOutput    *string  `json:"engine_output,omitempty"`
	Transmission    string   `json:"transmission" binding:"required,oneof=automatic manual"`
	FuelType        string   `json:"fuel_type" binding:"required,oneof=petrol diesel electric hybrid"`
	SeatingCapacity int      `json:"seating_capacity" binding:"required,min=2,max=15"`
	MaximumSpeed    *int     `json:"maximum_speed,omitempty"`
	Mileage         int      `json:"mileage" binding:"min=0"`
	DriverName      *string  `json:"driver_name,omitempty"`
	DriverNumber    *string  `json:"driver_number,omitempty"`
	DriverMiles     *int     `json:"driver_miles,omitempty"`
	HourlyRate      JSONB    `json:"hourly_rate" binding:"required"`
	CautionFee      float64  `json:"caution_fee" binding:"required,min=0"`
	Features        []string `json:"features,omitempty"`
	CurrentLocation *string  `json:"current_location,omitempty"`
}


type UpdateCarRequest struct {
	Model           *string  `json:"model,omitempty"`
	Brand           *string  `json:"brand,omitempty"`
	Year            *int     `json:"year,omitempty"`
	Color           *string  `json:"color,omitempty"`
	LicensePlate    *string  `json:"license_plate,omitempty"`
	EngineOutput    *string  `json:"engine_output,omitempty"`
	Transmission    *string  `json:"transmission,omitempty"`
	FuelType        *string  `json:"fuel_type,omitempty"`
	SeatingCapacity *int     `json:"seating_capacity,omitempty"`
	MaximumSpeed    *int     `json:"maximum_speed,omitempty"`
	Mileage         *int     `json:"mileage,omitempty"`
	DriverName      *string  `json:"driver_name,omitempty"`
	DriverNumber    *string  `json:"driver_number,omitempty"`
	DriverMiles     *int     `json:"driver_miles,omitempty"`
	HourlyRate      JSONB    `json:"hourly_rate,omitempty"`
	CautionFee      *float64 `json:"caution_fee,omitempty"`
	Features        []string `json:"features,omitempty"`
	IsAvailable     *bool    `json:"is_available,omitempty"`
	Status          *string  `json:"status,omitempty"`
	CurrentLocation *string  `json:"current_location,omitempty"`
}


type SearchCarsRequest struct {

	Brand           *string  `form:"brand"`
	Model           *string  `form:"model"`
	MinYear         *int     `form:"min_year"`
	MaxYear         *int     `form:"max_year"`
	Color           *string  `form:"color"`
	Transmission    *string  `form:"transmission"` // "automatic" or "manual"
	FuelType        *string  `form:"fuel_type"`
	MinSeats        *int     `form:"min_seats"`
	MaxSeats        *int     `form:"max_seats"`
	MinHourlyRate   *float64 `form:"min_hourly_rate"`
	MaxHourlyRate   *float64 `form:"max_hourly_rate"`
	Features        []string `form:"features"` // Must have these features
	Location        *string  `form:"location"`
	IsAvailable     *bool    `form:"is_available"`


	PickupDate  *time.Time `form:"pickup_date"`
	ReturnDate  *time.Time `form:"return_date"`


	SortBy  string `form:"sort_by" binding:"omitempty,oneof=hourly_rate year mileage seating_capacity"` // What to sort by
	OrderBy string `form:"order_by" binding:"omitempty,oneof=asc desc"` // asc or desc

	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=100"`
}


type CarListResponse struct {
	Cars       []*Car           		`json:"cars"`
	Pagination PaginationMeta   		`json:"pagination"`
	Filters    map[string]interface{} 	`json:"filters_applied"`
}


type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}