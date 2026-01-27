package models

type Car struct {
	ID string `json:"id" db:"id"`

	Model 			string  `json:"model" db:"model"`
	Brand			string  `json:"brand" db:"brand"`
	Year			int		`json:"year"  db:"year"`
	Color   		string	`json:"color" db:"color"`
	LicensePlace	string	`json:"licence_place" db:"license_plate"`

	// specification
	EngineOutput     *string `json:"engine_output,omitempty" db:"engine_output"` 
	Transmission	string	 `json:"transmission" db:"transmission"`
	FuelType		string   `json:"fuel_type" db:"fuel_type"`
	SeatingCapacity int      `json:"seating_capacity" db:"seating_capacity"`
	MaximumSpeed   *int      `json:"maximum_speed,omitempty" db:"maximum_speed"`
	Mileage			int		 `json:`

		// Driver Information (assigned driver)
	DriverName   *string `json:"driver_name,omitempty" db:"driver_name"`     
	DriverNumber *string `json:"driver_number,omitempty" db:"driver_number"` 
	DriverMiles  *int    `json:"driver_miles,omitempty" db:"driver_miles"`   



}